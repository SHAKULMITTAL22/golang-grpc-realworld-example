package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

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

