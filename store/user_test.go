package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
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
ROOST_METHOD_HASH=GetByEmail_3574af40e5
ROOST_METHOD_SIG_HASH=GetByEmail_5731b833c1


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
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func (s *UserStore) GetByID(id uint) (*model.User, error) {
	var m model.User
	if err := s.db.Find(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

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

/*
ROOST_METHOD_HASH=GetByUsername_f11f114df2
ROOST_METHOD_SIG_HASH=GetByUsername_954d096e24


 */
func TestGetByUsername(t *testing.T) {

	tests := []struct {
		name     string
		username string
		mock     func(mock sqlmock.Sqlmock)
		expected *model.User
		err      error
	}{
		{
			name:     "Retrieve User Successfully",
			username: "existing_user",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email"}).
					AddRow(1, "existing_user", "user@example.com")
				mock.ExpectQuery("^SELECT (.+) FROM (.+)$").WithArgs("existing_user").WillReturnRows(rows)
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "existing_user",
				Email:    "user@example.com",
			},
			err: nil,
		},
		{
			name:     "User Not Found",
			username: "non_existent_user",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM (.+)$").WithArgs("non_existent_user").WillReturnError(gorm.ErrRecordNotFound)
			},
			expected: nil,
			err:      gorm.ErrRecordNotFound,
		},
		{
			name:     "Database Connection Error",
			username: "any_user",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM (.+)$").WithArgs("any_user").WillReturnError(errors.New("db connection error"))
			},
			expected: nil,
			err:      errors.New("db connection error"),
		},
		{
			name:     "Multiple Users with Same Username",
			username: "duplicate_user",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email"}).
					AddRow(1, "duplicate_user", "first@example.com").
					AddRow(2, "duplicate_user", "second@example.com")
				mock.ExpectQuery("^SELECT (.+) FROM (.+)$").WithArgs("duplicate_user").WillReturnRows(rows)
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "duplicate_user",
				Email:    "first@example.com",
			},
			err: nil,
		},
		{
			name:     "Empty Username Input",
			username: "",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM (.+)$").WithArgs("").WillReturnError(errors.New("invalid input syntax"))
			},
			expected: nil,
			err:      errors.New("invalid input syntax"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.mock(mock)

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
			}

			userStore := &UserStore{db: gormDB}

			user, err := userStore.GetByUsername(tt.username)

			if user != nil && tt.expected != nil {
				if user.ID != tt.expected.ID || user.Username != tt.expected.Username || user.Email != tt.expected.Email {
					t.Errorf("unexpected user data, got: %+v, want: %+v", user, tt.expected)
				}
			} else if user != tt.expected {
				t.Errorf("unexpected user, got: %+v, want: %+v", user, tt.expected)
			}

			if (err != nil && tt.err == nil) || (err == nil && tt.err != nil) || (err != nil && tt.err != nil && err.Error() != tt.err.Error()) {
				t.Errorf("unexpected error, got: %v, want: %v", err, tt.err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			t.Logf("Test '%s' completed successfully", tt.name)
		})
	}
}

/*
ROOST_METHOD_HASH=IsFollowing_f53a5d9cef
ROOST_METHOD_SIG_HASH=IsFollowing_9eba5a0e9c


 */
func TestUserStoreIsFollowing(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a Gorm DB connection", err)
	}

	userStore := &UserStore{db: gormDB}

	type testCase struct {
		name       string
		setupMocks func()
		userA      *model.User
		userB      *model.User
		expected   bool
		expectErr  bool
	}

	tests := []testCase{
		{
			name: "Valid follow relationship",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT count(.+) FROM follows WHERE from_user_id = \$1 AND to_user_id = \$2`).
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			userA:     &model.User{Model: gorm.Model{ID: 1}},
			userB:     &model.User{Model: gorm.Model{ID: 2}},
			expected:  true,
			expectErr: false,
		},
		{
			name: "Non-existing follow relationship",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT count(.+) FROM follows WHERE from_user_id = \$1 AND to_user_id = \$2`).
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			userA:     &model.User{Model: gorm.Model{ID: 1}},
			userB:     &model.User{Model: gorm.Model{ID: 2}},
			expected:  false,
			expectErr: false,
		},
		{
			name: "Nil user input",
			setupMocks: func() {

			},
			userA:     nil,
			userB:     &model.User{Model: gorm.Model{ID: 2}},
			expected:  false,
			expectErr: false,
		},
		{
			name: "Database error handling",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT count(.+) FROM follows WHERE from_user_id = \$1 AND to_user_id = \$2`).
					WithArgs(1, 2).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			userA:     &model.User{Model: gorm.Model{ID: 1}},
			userB:     &model.User{Model: gorm.Model{ID: 2}},
			expected:  false,
			expectErr: true,
		},
		{
			name: "Self follow check",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT count(.+) FROM follows WHERE from_user_id = \$1 AND to_user_id = \$2`).
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			userA:     &model.User{Model: gorm.Model{ID: 1}},
			userB:     &model.User{Model: gorm.Model{ID: 1}},
			expected:  false,
			expectErr: false,
		},
		{
			name: "Multiple follow relationships",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT count(.+) FROM follows WHERE from_user_id = \$1 AND to_user_id = \$2`).
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			userA:     &model.User{Model: gorm.Model{ID: 1}},
			userB:     &model.User{Model: gorm.Model{ID: 2}},
			expected:  true,
			expectErr: false,
		},
		{
			name: "Empty database",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT count(.+) FROM follows WHERE from_user_id = \$1 AND to_user_id = \$2`).
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			userA:     &model.User{Model: gorm.Model{ID: 1}},
			userB:     &model.User{Model: gorm.Model{ID: 2}},
			expected:  false,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			result, err := userStore.IsFollowing(tt.userA, tt.userB)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

