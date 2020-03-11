package storage

import (
	"time"

	"github.com/jmoiron/sqlx"
	//"github.com/pkg/errors"
	"testing"

	_ "github.com/lib/pq"
	mgr "github.com/rubenv/sql-migrate"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUser(t *testing.T) {
	db, err := sqlx.Open("postgres", "postgres://postgres:yj12345@2.59.151.166:1932/postgres?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	for {
		if err := db.Ping(); err != nil {
			t.Errorf("ping database error, will retry in 2s: %s", err)
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	Convey("Given a clean database", t, func() {
		m := &mgr.FileMigrationSource{
			Dir: "/home/yjiong/GOPATH/src/github.com/yjiong/fkhj212/migrations/",
		}
		if _, err := mgr.Exec(db.DB, "postgres", m, mgr.Down); err != nil {
			t.Error(err)
		}
		if _, err := mgr.Exec(db.DB, "postgres", m, mgr.Up); err != nil {
			t.Error(err)
		}
		So(err, ShouldBeNil)
		Convey("When creating a user with an invalid username", func() {
			user := User{
				Username: "bad characters %",
			}
			_, err := CreateUser(db, &user, "somepassword")
			So(err, ShouldNotBeNil)
		})
		Convey("When creating a user", func() {
			user := User{
				Username:   "goodusername111",
				IsAdmin:    false,
				SessionTTL: 20,
			}
			password := "somepassword"

			userID, err := CreateUser(db, &user, password)
			t.Log("userID", userID)
			So(err, ShouldBeNil)

			Convey("It can be get by id", func() {
				user2, _ := GetUser(db, userID)
				t.Log(user2)
				So(user2.Username, ShouldResemble, user.Username)
				So(user2.IsAdmin, ShouldResemble, user.IsAdmin)
				So(user2.SessionTTL, ShouldResemble, user.SessionTTL)

				Convey("It can be get by username", func() {
					user2, err := GetUserByUsername(db, user.Username)
					So(err, ShouldBeNil)
					So(user2.Username, ShouldResemble, user.Username)
					So(user2.IsAdmin, ShouldResemble, user.IsAdmin)
					So(user2.SessionTTL, ShouldResemble, user.SessionTTL)
				})
			})
		})
	})

	/*
			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Cause(err), ShouldResemble, ErrUserInvalidUsername)
			})
			//})

			Convey("When creating a user with an invalid password", func() {
				user := User{
					Username:   "okcharacters",
					IsAdmin:    false,
					SessionTTL: 40,
					Email:      "foo@bar.com",
				}
				_, err := CreateUser(db, &user, "bad")

				Convey("Then an error is returned", func() {
					So(err, ShouldNotBeNil)
					So(errors.Cause(err), ShouldResemble, ErrUserPasswordLength)
				})
			})

			Convey("When creating a user with an invalid e-mail", func() {
				user := User{
					Username:   "okcharacters",
					IsAdmin:    false,
					SessionTTL: 40,
					Email:      "foobar.com",
				}
				_, err := CreateUser(db, &user, "somepassword")

				Convey("Then an error is returned", func() {
					So(err, ShouldNotBeNil)
					So(errors.Cause(err), ShouldResemble, ErrInvalidEmail)
				})
			})

			Convey("When creating a user", func() {
				user := User{
					Username:   "goodusername111",
					IsAdmin:    false,
					SessionTTL: 20,
					Email:      "foo@bar.com",
				}
				password := "somepassword"

				userID, err := CreateUser(db, &user, password)
				So(err, ShouldBeNil)

				Convey("It can be get by id", func() {
					user2, err := GetUser(db, userID)
					So(err, ShouldBeNil)
					So(user2.Username, ShouldResemble, user.Username)
					So(user2.IsAdmin, ShouldResemble, user.IsAdmin)
					So(user2.SessionTTL, ShouldResemble, user.SessionTTL)
				})

				Convey("It can be get by username", func() {
					user2, err := GetUserByUsername(db, user.Username)
					So(err, ShouldBeNil)
					So(user2.Username, ShouldResemble, user.Username)
					So(user2.IsAdmin, ShouldResemble, user.IsAdmin)
					So(user2.SessionTTL, ShouldResemble, user.SessionTTL)
				})

				Convey("Then get users returns 2 users", func() {
					users, err := GetUsers(db, 10, 0, "")
					So(err, ShouldBeNil)
					So(users, ShouldHaveLength, 2)
					checkUser := 0
					if users[0].Username == "admin" {
						// No, check entry 1
						checkUser = 1
					}
					So(users[checkUser].Username, ShouldResemble, user.Username)
					So(users[checkUser].IsAdmin, ShouldResemble, user.IsAdmin)
					So(users[checkUser].SessionTTL, ShouldResemble, user.SessionTTL)
				})

				Convey("Then get user count returns 2", func() {
					count, err := GetUserCount(db, "")
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 2)
				})

				Convey("Then searching for 'good' returns a single item", func() {
					count, err := GetUserCount(db, "good")
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 1)

					users, err := GetUsers(db, 10, 0, "good")
					So(err, ShouldBeNil)
					So(users, ShouldHaveLength, 1)
				})

				Convey("Then searching for 'foo' returns 0 items", func() {
					count, err := GetUserCount(db, "foo")
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)

					users, err := GetUsers(db, 10, 0, "foo")
					So(err, ShouldBeNil)
					So(users, ShouldHaveLength, 0)
				})

				Convey("Then the user can t in", func() {
					jwt, err := tinUser(db, user.Username, password)
					So(err, ShouldBeNil)
					So(jwt, ShouldNotBeNil)
				})

				Convey("When updating the user password", func() {
					password = "newrandompassword2*&^"
					So(UpdatePassword(db, user.ID, password), ShouldBeNil)

					Convey("Then the user can t in with the new password", func() {
						jwt, err := tinUser(db, user.Username, password)
						So(err, ShouldBeNil)
						So(jwt, ShouldNotBeNil)
					})
				})

				Convey("When updating the user", func() {
					userUpdate := UserUpdate{
						ID:         user.ID,
						Username:   "newusername",
						IsAdmin:    true,
						SessionTTL: 30,
						Email:      "bar@foo.com",
					}
					So(UpdateUser(db, userUpdate), ShouldBeNil)

					Convey("Then the user has been updated", func() {
						user2, err := GetUser(db, user.ID)
						So(err, ShouldBeNil)
						So(user2.Username, ShouldResemble, userUpdate.Username)
						So(user2.IsAdmin, ShouldResemble, userUpdate.IsAdmin)
						So(user2.SessionTTL, ShouldResemble, userUpdate.SessionTTL)
					})
				})

				Convey("When deleting the user", func() {
					So(DeleteUser(db, user.ID), ShouldBeNil)

					Convey("Then the user count returns 1", func() {
						count, err := GetUserCount(db, "")
						So(err, ShouldBeNil)
						So(count, ShouldEqual, 1)
					})
				})

			})
		})
	*/
}
