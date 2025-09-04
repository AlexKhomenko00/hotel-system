package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/AlexKhomenko00/hotel-system/internal/auth"
	"github.com/AlexKhomenko00/hotel-system/internal/auth/jwt"
	"github.com/AlexKhomenko00/hotel-system/internal/config"
	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testSuite *TestSuite
	once      sync.Once
)

func GetTestSuite() *TestSuite {
	once.Do(func() {
		validator := validator.New()

		testConfig := &config.Config{
			JWTSecret:          "test-jwt-secret-key",
			PORT:               "8080",
			DB_HOST:            "localhost",
			DB_PORT:            "5432",
			DB_DATABASE:        "hotel_test",
			DB_USERNAME:        "test",
			DB_PASSWORD:        "test",
			DB_SCHEMA:          "public",
			DB_SSLMODE:         "disable",
			OVERBOOKING_FACTOR: "1.2",
		}

		testSuite = &TestSuite{
			ctx:       context.Background(),
			config:    testConfig,
			validator: validator,
		}

		if err := testSuite.SetupApplication(); err != nil {
			log.Fatalf("Failed to setup test application: %v", err)
		}
	})
	return testSuite
}

type TestSuite struct {
	server      *httptest.Server
	handler     http.Handler
	db          database.Service
	queries     *database.Queries
	auth        jwt.Authenticator
	config      *config.Config
	validator   *validator.Validate
	ctx         context.Context
	pgContainer *postgres.PostgresContainer
	r           *chi.Mux
	authSvc     *auth.AuthService
}

func TestMain(m *testing.M) {
	suite := GetTestSuite()

	code := m.Run()

	suite.TearDown()
	os.Exit(code)
}

func (ts *TestSuite) SetupApplication() error {
	var err error

	ts.pgContainer, err = postgres.Run(ts.ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("hotel_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute)),
	)
	if err != nil {
		return fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := ts.pgContainer.ConnectionString(ts.ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	mappedPort, err := ts.pgContainer.MappedPort(ts.ctx, "5432")
	if err != nil {
		return fmt.Errorf("failed to get mapped port: %w", err)
	}
	host, err := ts.pgContainer.Host(ts.ctx)
	if err != nil {
		return fmt.Errorf("failed to get container host: %w", err)
	}

	ts.config.DB_HOST = host
	ts.config.DB_PORT = mappedPort.Port()

	if err := ts.runMigrations(connStr); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	ts.db, err = database.Create(ts.config)
	if err != nil {
		return fmt.Errorf("failed to create database service: %w", err)
	}

	ts.queries = database.New(ts.db.GetDB())
	ts.auth = jwt.NewAuthenticator(ts.config.JWTSecret)

	ts.handler = ts.createTestHandler()
	ts.authSvc = auth.New(ts.queries, ts.validator, ts.config)

	log.Printf("Test database connection: %s", connStr)
	return nil
}

func (ts *TestSuite) runMigrations(connStr string) error {
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	migrationsPath := filepath.Join(projectRoot, "sql", "schema")

	migrationDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open migration connection: %w", err)
	}
	defer migrationDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(migrationDB, migrationsPath); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}

	return nil
}

func (ts *TestSuite) TearDown() {
	if ts.server != nil {
		ts.server.Close()
	}
	if ts.db != nil {
		ts.db.Close()
	}
	if ts.pgContainer != nil {
		if err := ts.pgContainer.Terminate(ts.ctx); err != nil {
			log.Printf("Warning: failed to terminate postgres container: %v", err)
		}
	}
}

func (ts *TestSuite) GetServerURL() string {
	if ts.server == nil {
		return "http://localhost:8080"
	}
	return ts.server.URL
}

func (ts *TestSuite) GetDB() database.Service {
	return ts.db
}

func (ts *TestSuite) GetQueries() *database.Queries {
	return ts.queries
}

func (ts *TestSuite) GetConfig() config.Config {
	return *ts.config
}

func (ts *TestSuite) GetValidator() *validator.Validate {
	return ts.validator
}

func (ts *TestSuite) CleanupDatabase() error {
	queries := []string{
		"TRUNCATE TABLE booking.reservations CASCADE",
		"TRUNCATE TABLE booking.room_type_inventory CASCADE",
		"TRUNCATE TABLE booking.rooms CASCADE",
		"TRUNCATE TABLE booking.room_types CASCADE",
		"TRUNCATE TABLE booking.hotels CASCADE",
		"TRUNCATE TABLE booking.guests CASCADE",
		"TRUNCATE TABLE auth.users CASCADE",
	}

	for _, query := range queries {
		if _, err := ts.db.GetDB().Exec(query); err != nil {
			log.Printf("Warning: failed to cleanup table with query %s: %v", query, err)
		}
	}

	return nil
}

func (ts *TestSuite) GenerateJWT(usr database.AuthUser) (string, error) {
	_, tokenString, err := ts.auth.EncodeUserClaims(usr)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return tokenString, nil
}

func (ts *TestSuite) CreateTestUser() (database.AuthUser, error) {
	guest, err := ts.CreateTestGuest()
	if err != nil {
		return database.AuthUser{}, fmt.Errorf("failed to create guest for user: %w", err)
	}

	userID := uuid.New()
	user := database.InsertUserParams{
		ID:           userID,
		Email:        fmt.Sprintf("test_%s@example.com", userID.String()[:8]),
		PasswordHash: "hashedpassword123",
		GuestID:      guest.ID,
	}

	createdUser, err := ts.queries.InsertUser(ts.ctx, user)
	if err != nil {
		return database.AuthUser{}, fmt.Errorf("failed to create test user: %w", err)
	}

	return createdUser, nil
}

func (ts *TestSuite) CreateTestHotel() (database.BookingHotel, error) {
	hotelID := uuid.New()
	hotel := database.CreateHotelParams{
		ID:       hotelID,
		Name:     fmt.Sprintf("Test Hotel %s", hotelID.String()[:8]),
		Location: "Greece/Athens",
	}

	createdHotel, err := ts.queries.CreateHotel(ts.ctx, hotel)
	if err != nil {
		return database.BookingHotel{}, fmt.Errorf("failed to create test hotel: %w", err)
	}

	return createdHotel, nil
}

func (ts *TestSuite) CreateTestRoomType(hotelID uuid.UUID) (database.BookingRoomType, error) {
	roomTypeID := uuid.New()
	roomType := database.CreateRoomTypeParams{
		ID:          roomTypeID,
		HotelID:     hotelID,
		Name:        fmt.Sprintf("Standard_%s", roomTypeID.String()[:8]),
		Description: sql.NullString{String: "Standard room type for testing", Valid: true},
	}

	createdRoomType, err := ts.queries.CreateRoomType(ts.ctx, roomType)
	if err != nil {
		return database.BookingRoomType{}, fmt.Errorf("failed to create test room type: %w", err)
	}

	return createdRoomType, nil
}

func (ts *TestSuite) CreateTestRoom(hotelID uuid.UUID, roomTypeID uuid.UUID) (database.BookingRoom, error) {
	roomID := uuid.New()
	room := database.CreateRoomParams{
		ID:         roomID,
		HotelID:    hotelID,
		Name:       fmt.Sprintf("Room_%d", time.Now().Unix()%1000),
		Floor:      1,
		Number:     int32(time.Now().Unix() % 1000),
		RoomTypeID: roomTypeID,
		Status:     database.BookingRoomStatusAvailable,
	}

	createdRoom, err := ts.queries.CreateRoom(ts.ctx, room)
	if err != nil {
		return database.BookingRoom{}, fmt.Errorf("failed to create test room: %w", err)
	}

	return createdRoom, nil
}

func (ts *TestSuite) CreateTestGuest() (database.BookingGuest, error) {
	guestID := uuid.New()
	_, err := ts.db.GetDB().Exec(`
		INSERT INTO booking.guests (id, first_name, last_name, email) 
		VALUES ($1, $2, $3, $4)`,
		guestID, "TestFirst", "TestLast", fmt.Sprintf("guest_%s@example.com", guestID.String()[:8]))

	if err != nil {
		return database.BookingGuest{}, fmt.Errorf("failed to create test guest: %w", err)
	}

	return database.BookingGuest{
		ID:        guestID,
		FirstName: "TestFirst",
		LastName:  "TestLast",
		Email:     fmt.Sprintf("guest_%s@example.com", guestID.String()[:8]),
	}, nil
}

func (ts *TestSuite) CreateTestReservation(guestID uuid.UUID, roomTypeID uuid.UUID, hotelID uuid.UUID, startDate, endDate time.Time) (database.BookingReservation, error) {
	reservationID := uuid.New()
	_, err := ts.db.GetDB().Exec(`
		INSERT INTO booking.reservations (id, hotel_id, room_type_id, start_date, end_date, status, guest_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		reservationID, hotelID, roomTypeID, startDate, endDate, "confirmed", guestID)

	if err != nil {
		return database.BookingReservation{}, fmt.Errorf("failed to create test reservation: %w", err)
	}

	return database.BookingReservation{
		ID:         reservationID,
		HotelID:    hotelID,
		RoomTypeID: roomTypeID,
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     "confirmed",
		GuestID:    guestID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

type TestData struct {
	User        database.AuthUser
	Hotel       database.BookingHotel
	RoomType    database.BookingRoomType
	Room        database.BookingRoom
	Guest       database.BookingGuest
	Reservation database.BookingReservation
	JWT         string
}

func (ts *TestSuite) CreateFullTestData() (*TestData, error) {
	user, err := ts.CreateTestUser()
	if err != nil {
		return nil, fmt.Errorf("failed to create test user: %w", err)
	}

	jwt, err := ts.GenerateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	hotel, err := ts.CreateTestHotel()
	if err != nil {
		return nil, fmt.Errorf("failed to create test hotel: %w", err)
	}

	roomType, err := ts.CreateTestRoomType(hotel.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create test room type: %w", err)
	}

	room, err := ts.CreateTestRoom(hotel.ID, roomType.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create test room: %w", err)
	}

	guest, err := ts.CreateTestGuest()
	if err != nil {
		return nil, fmt.Errorf("failed to create test guest: %w", err)
	}

	checkIn := time.Now().AddDate(0, 0, 1)  // Tomorrow
	checkOut := time.Now().AddDate(0, 0, 2) // Day after tomorrow

	reservation, err := ts.CreateTestReservation(guest.ID, roomType.ID, hotel.ID, checkIn, checkOut)
	if err != nil {
		return nil, fmt.Errorf("failed to create test reservation: %w", err)
	}

	return &TestData{
		User:        user,
		Hotel:       hotel,
		RoomType:    roomType,
		Room:        room,
		Guest:       guest,
		Reservation: reservation,
		JWT:         jwt,
	}, nil
}

func (ts *TestSuite) createTestHandler() http.Handler {
	ts.r = chi.NewRouter()
	ts.r.Use(middleware.Logger)

	ts.r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	return ts.r
}

func (ts *TestSuite) RegisterPrivateHandlers(registerHandlers func(r chi.Router), path string) {
	ts.r.Group(func(r chi.Router) {
		ts.authSvc.SetupJWTAuthMiddleware(r)

		r.Route(path, func(r chi.Router) {
			registerHandlers(r)
		})
	})
}

func (ts *TestSuite) GetHandler() http.Handler {
	return ts.handler
}

func (ts *TestSuite) MakeAuthenticatedRequest(method, url string, body any, user database.AuthUser) (*http.Response, error) {
	jwt, err := ts.GenerateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, url, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	fmt.Printf("fucking JWT %s \n", jwt)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

	w := httptest.NewRecorder()
	ts.handler.ServeHTTP(w, req)

	return w.Result(), nil
}

func (ts *TestSuite) MakeRequest(method, url string, body any) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, url, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	ts.handler.ServeHTTP(w, req)

	return w.Result(), nil
}
