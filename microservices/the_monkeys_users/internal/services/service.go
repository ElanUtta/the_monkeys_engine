package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/the-monkeys/the_monkeys/apis/serviceconn/gateway_user/pb"
	"github.com/the-monkeys/the_monkeys/constants"
	"github.com/the-monkeys/the_monkeys/microservices/the_monkeys_users/internal/database"
	"github.com/the-monkeys/the_monkeys/microservices/the_monkeys_users/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserSvc struct {
	dbConn database.UserDb
	log    *logrus.Logger
	pb.UnimplementedUserServiceServer
}

func NewUserSvc(dbConn database.UserDb, log *logrus.Logger) *UserSvc {
	return &UserSvc{
		dbConn: dbConn,
		log:    log,
	}
}

func (us *UserSvc) GetUserProfile(ctx context.Context, req *pb.UserProfileReq) (*pb.UserProfileRes, error) {
	us.log.Infof("user %v has requested profile info.", req.Username)
	if !req.IsPrivate {
		userProfile, err := us.dbConn.GetUserProfile(req.Username)
		if err != nil {
			us.log.Errorf("the user doesn't exists: %v", err)
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("user %s doesn't exist: %v", req.Username, err))
		}
		return &pb.UserProfileRes{
			Username:  userProfile.UserName,
			FirstName: userProfile.FirstName,
			LastName:  userProfile.LastName,
			Bio:       userProfile.Bio.String,
			AvatarUrl: userProfile.AvatarUrl.String,
		}, nil

	}

	_, err := us.dbConn.CheckIfUsernameExist(req.Username)
	if err != nil {
		us.log.Errorf("the user doesn't exists: %v", err)
		return nil, err
	}

	userDetails, err := us.dbConn.GetMyProfile(req.Username)
	if err != nil {
		us.log.Errorf("error while finding the user profile: %v", err)
		return nil, err
	}

	// us.log.Infof("get profile: userDetails, %+v", userDetails)
	return &pb.UserProfileRes{
		AccountId:   userDetails.AccountId,
		Username:    userDetails.Username,
		FirstName:   userDetails.FirstName,
		LastName:    userDetails.LastName,
		DateOfBirth: userDetails.DateOfBirth.Time.String(),
		Bio:         userDetails.Bio.String,
		AvatarUrl:   userDetails.AvatarUrl.String,
		// CreatedAt:     userDetails.CreatedAt.,
		// UpdatedAt:     userDetails.UpdatedAt,
		Address: userDetails.Address.String,
		// ContactNumber: userDetails.ContactNumber.String,
		UserStatus: userDetails.UserStatus,
	}, err
}

func (us *UserSvc) GetUserActivities(ctx context.Context, req *pb.UserActivityReq) (*pb.UserActivityRes, error) {
	logrus.Infof("Trying to fetch user activities for: %v", req.Email)

	return &pb.UserActivityRes{}, nil
}
func (us *UserSvc) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileReq) (*pb.UpdateUserProfileRes, error) {
	us.log.Infof("user %s is updating the profile.", req.Username)

	// Check if the user exists
	_, err := us.dbConn.CheckIfUsernameExist(req.Username)
	if err != nil {
		us.log.Errorf("the user doesn't exists: %v", err)
		return nil, err
	}

	// Check if the method isPartial true
	var dbUserInfo *models.UserProfileRes
	if req.Partial {
		// If isPartial is true fetch the remaining data from the db
		dbUserInfo, err = us.dbConn.GetMyProfile(req.Username)
		if err != nil {
			us.log.Errorf("error while finding the user profile: %v", err)
			return nil, err
		}
	}

	// Map the user
	mappedDBUser := MapUserUpdateData(req, dbUserInfo)
	if err != nil {
		return nil, err
	}

	us.log.Infof("mappedDBUser: %+v\n", mappedDBUser)
	// Update the user
	err = us.dbConn.UpdateUserProfile(req.Username, mappedDBUser)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateUserProfileRes{
		Username: mappedDBUser.Username,
	}, err
}

// MapUserUpdateData maps the user update request data to the database model.
func MapUserUpdateData(req *pb.UpdateUserProfileReq, dbUserInfo *models.UserProfileRes) *models.UserProfileRes {
	if req.Username != "" {
		dbUserInfo.Username = req.Username
	}
	if req.FirstName != "" {
		dbUserInfo.FirstName = req.FirstName
	}
	if req.LastName != "" {
		dbUserInfo.LastName = req.LastName
	}
	if req.Bio != "" {
		dbUserInfo.Bio.String = req.Bio
	}
	if req.DateOfBirth != "" {
		time, _ := time.Parse(constants.DateTimeFormat, req.DateOfBirth)
		dbUserInfo.DateOfBirth.Time = time
	}
	if req.Address != "" {
		dbUserInfo.Address.String = req.Address
	}
	if req.ContactNumber != "0" {
		dbUserInfo.ContactNumber.String = req.ContactNumber
	}

	return dbUserInfo
}
func (us *UserSvc) DeleteUserProfile(ctx context.Context, req *pb.DeleteUserProfileReq) (*pb.DeleteUserProfileRes, error) {
	us.log.Infof("user %s has requested to delete the  profile.", req.Username)

	// Check if username exits or not
	_, err := us.dbConn.CheckIfUsernameExist(req.Username)
	if err != nil {
		us.log.Errorf("the user doesn't exists: %v", err)
		return nil, err
	}

	// Run delete user query
	err = us.dbConn.DeleteUserProfile(req.Username)
	if err != nil {
		us.log.Errorf("could not delete the user profile: %v", err)
		return nil, err
	}

	// Return the response
	return &pb.DeleteUserProfileRes{
		Success: "user has been deleted successfully",
		Status:  "200",
	}, nil

}

func (us *UserSvc) GetAllTopics(context.Context, *pb.GetTopicsRequests) (*pb.GetTopicsResponse, error) {
	us.log.Info("getting all the topics")

	res, err := us.dbConn.GetAllTopicsFromDb()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			us.log.Errorf("cannot find the topics in the database: %v", err)
		}
		us.log.Errorf("error while querrying the topics: %v", err)
		return nil, errors.New("error while querrying the topics")
	}

	return res, err
}

func (us *UserSvc) GetAllCategories(ctx context.Context, req *pb.GetAllCategoriesReq) (*pb.GetAllCategoriesRes, error) {
	us.log.Info("getting all the Description and Categories")

	res, err := us.dbConn.GetAllCategories()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			us.log.Errorf("no Categories and Description found in the database: %v", err)
			return nil, errors.New("no Categories found")
		}
		us.log.Errorf("error while querying the Categories: %v", err)
		return nil, errors.New("error while querying the categories")
	}

	return res, nil
}
