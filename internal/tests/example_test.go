package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleHotelCreation(t *testing.T) {
	suite := GetTestSuite()

	err := suite.CleanupDatabase()
	require.NoError(t, err, "Failed to cleanup database")

	hotel, err := suite.CreateTestHotel()
	require.NoError(t, err, "Failed to create test hotel")
	assert.NotEmpty(t, hotel.ID, "Hotel ID should not be empty")
	assert.Contains(t, hotel.Name, "Test Hotel", "Hotel name should contain 'Test Hotel'")
	assert.Equal(t, "Greece/Athens", hotel.Location, "Hotel location should be Greece/Athens")
	assert.True(t, hotel.IsActive, "Hotel should be active by default")
}

func TestExampleFullTestData(t *testing.T) {
	suite := GetTestSuite()

	err := suite.CleanupDatabase()
	require.NoError(t, err, "Failed to cleanup database")

	testData, err := suite.CreateFullTestData()
	require.NoError(t, err, "Failed to create full test data")

	assert.NotEmpty(t, testData.User.ID, "User ID should not be empty")
	assert.Contains(t, testData.User.Email, "@example.com", "User email should contain @example.com")
	assert.NotEmpty(t, testData.JWT, "JWT should not be empty")

	assert.NotEmpty(t, testData.Hotel.ID, "Hotel ID should not be empty")
	assert.Contains(t, testData.Hotel.Name, "Test Hotel", "Hotel name should contain 'Test Hotel'")

	assert.NotEmpty(t, testData.RoomType.ID, "Room type ID should not be empty")
	assert.Equal(t, testData.Hotel.ID, testData.RoomType.HotelID, "Room type should belong to the hotel")
	assert.Contains(t, testData.RoomType.Name, "Standard_", "Room type name should contain 'Standard_'")

	assert.NotEmpty(t, testData.Room.ID, "Room ID should not be empty")
	assert.Equal(t, testData.Hotel.ID, testData.Room.HotelID, "Room should belong to the hotel")
	assert.Equal(t, testData.RoomType.ID, testData.Room.RoomTypeID, "Room should have the correct room type")

	assert.NotEmpty(t, testData.Guest.ID, "Guest ID should not be empty")
	assert.Equal(t, "TestFirst", testData.Guest.FirstName, "Guest first name should be TestFirst")
	assert.Equal(t, "TestLast", testData.Guest.LastName, "Guest last name should be TestLast")
	assert.Contains(t, testData.Guest.Email, "@example.com", "Guest email should contain @example.com")

	assert.NotEmpty(t, testData.Reservation.ID, "Reservation ID should not be empty")
	assert.Equal(t, testData.Hotel.ID, testData.Reservation.HotelID, "Reservation should belong to the hotel")
	assert.Equal(t, testData.RoomType.ID, testData.Reservation.RoomTypeID, "Reservation should be for the correct room type")
	assert.Equal(t, testData.Guest.ID, testData.Reservation.GuestID, "Reservation should belong to the guest")
	assert.Equal(t, "confirmed", testData.Reservation.Status, "Reservation status should be confirmed")
	assert.True(t, testData.Reservation.EndDate.After(testData.Reservation.StartDate), "End date should be after start date")
}

func TestExampleCleanupDatabase(t *testing.T) {
	suite := GetTestSuite()

	_, err := suite.CreateTestUser()
	require.NoError(t, err, "Failed to create test user")

	_, err = suite.CreateTestHotel()
	require.NoError(t, err, "Failed to create test hotel")

	err = suite.CleanupDatabase()
	require.NoError(t, err, "Failed to cleanup database")

	_, err = suite.CreateTestUser()
	require.NoError(t, err, "Failed to create test user after cleanup")
}
