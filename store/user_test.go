package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

type UserStore struct {
	db *gorm.DB
}
type DB struct {
	sync.RWMutex
	Value			interface{}
	Error			error
	RowsAffected		int64
	db			SQLCommon
	blockGlobalUpdate	bool
	logMode			logModeValue
	logger			logger
	search			*search
	values			sync.Map
	parent			*DB
	callbacks		*Callback
	dialect			Dialect
	singularTable		bool
	nowFuncOverride		func() time.Time
}// single db
// function to be used to override the creating of a new timestamp


/*
 */
func TestGetByEmail(t *testing.T) {

	tests := []struct {
		name         string
		email        string
		mockSetup    func(sqlmock.Sqlmock)
		expectedUser *model.User
		expectedErr  error
	}{
		{
			name:  "Retrieve User Successfully by Email",
			email: "test@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"email", "username"}).
					AddRow("test@example.com", "testuser")
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(email = \\?\\)").
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedUser: &model.User{Email: "test@example.com", Username: "testuser"},
			expectedErr:  nil,
		},
		{
			name:  "No User Found for Given Email",
			email: "nonexistent@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(email = \\?\\)").
					WithArgs("nonexistent@example.com").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			name:  "Database Error Occurs",
			email: "error@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(email = \\?\\)").
					WithArgs("error@example.com").
					WillReturnError(errors.New("database error"))
			},
			expectedUser: nil,
			expectedErr:  errors.New("database error"),
		},
		{
			name:  "Empty Email String Provided",
			email: "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(email = \\?\\)").
					WithArgs("").
					WillReturnError(errors.New("invalid query parameters"))
			},
			expectedUser: nil,
			expectedErr:  errors.New("invalid query parameters"),
		},
		{
			name:  "Special Characters in Email",
			email: "user+test@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"email", "username"}).
					AddRow("user+test@example.com", "specialuser")
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(email = \\?\\)").
					WithArgs("user+test@example.com").
					WillReturnRows(rows)
			},
			expectedUser: &model.User{Email: "user+test@example.com", Username: "specialuser"},
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}
			defer gormDB.Close()

			store := &UserStore{db: gormDB}

			tt.mockSetup(mock)

			user, err := store.GetByEmail(tt.email)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tt.expectedUser != nil {
				if user == nil {
					t.Errorf("expected user %v, got nil", tt.expectedUser)
				} else if user.Email != tt.expectedUser.Email || user.Username != tt.expectedUser.Username {
					t.Errorf("expected user %v, got %v", tt.expectedUser, user)
				}
			} else if user != nil {
				t.Errorf("expected nil user, got %v", user)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

/*
 */
func (s *UserStore) GetByID(id uint) (*model.User, error) {
	var m model.User
	if err := s.db.Find(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

/*
 */
func TestUserStoreGetByID(t *testing.T) {

	tests := []struct {
		name          string
		setupMock     func(sqlmock.Sqlmock)
		id            uint
		expectedUser  *model.User
		expectedError error
	}{
		{
			name: "Retrieve Existing User by ID",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE \\(id = \\?\\)").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "John Doe"))
			},
			id: 1,
			expectedUser: &model.User{
				ID:   1,
				Name: "John Doe",
			},
			expectedError: nil,
		},
		{
			name: "User Not Found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE \\(id = \\?\\)").
					WithArgs(2).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			id:            2,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Database Connection Error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE \\(id = \\?\\)").
					WithArgs(3).
					WillReturnError(errors.New("connection error"))
			},
			id:            3,
			expectedUser:  nil,
			expectedError: errors.New("connection error"),
		},
		{
			name: "Invalid ID Input",
			setupMock: func(mock sqlmock.Sqlmock) {

			},
			id:            0,
			expectedUser:  nil,
			expectedError: gorm.ErrInvalidSQL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open mock sql db, %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("failed to open gorm db, %v", err)
			}

			tt.setupMock(mock)

			store := &UserStore{db: gormDB}

			user, err := store.GetByID(tt.id)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tt.expectedUser != nil {
				if user == nil || user.ID != tt.expectedUser.ID || user.Name != tt.expectedUser.Name {
					t.Errorf("expected user: %v, got: %v", tt.expectedUser, user)
				}
			} else {
				if user != nil {
					t.Errorf("expected nil user, got: %v", user)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}

}

