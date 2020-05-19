package main

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	googleAPI "timecodes/cmd/google_api"
)

type UserRepositorySuite struct {
	suite.Suite
	DB   *gorm.DB
	Repo *DBUserRepository
}

func (suite *UserRepositorySuite) SetupSuite() {
	suite.DB = TestDB
	suite.Repo = &DBUserRepository{DB: TestDB}
}

func (suite *UserRepositorySuite) SetupTest() {
	Cleaner.Acquire("users")
}

func (suite *UserRepositorySuite) TearDownTest() {
	Cleaner.Clean("users")
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}

func (suite *UserRepositorySuite) TestDBUserRepository_FindOrCreateByGoogleInfo() {
	t := suite.T()

	t.Run("when user doesn't exist", func(t *testing.T) {
		googleID := "10001"
		userInfo := &googleAPI.UserInfo{ID: googleID}

		user := suite.Repo.FindOrCreateByGoogleInfo(userInfo)

		assert.Equal(t, "10001", user.GoogleID)
	})

	t.Run("when user exists", func(t *testing.T) {
		googleID := "10002"

		suite.DB.Create(&User{GoogleID: googleID})

		userInfo := &googleAPI.UserInfo{ID: googleID}

		user := suite.Repo.FindOrCreateByGoogleInfo(userInfo)

		assert.Equal(t, "10002", user.GoogleID)
	})
}
