package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	//"io/ioutil"

	"log"
	"time"

	"github.com/ar-maged/guc-api/factory"
	"github.com/ar-maged/guc-api/util"
	cors "github.com/heppu/simple-cors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

var (
	// WelcomeMessage A constant to hold the welcome message
	WelcomeMessage = "Hello, GUC student"

	// sessions = {
	//   "uuid1" = Session{...},
	//   ...
	// }
	sessions = map[string]Session{}
	DB       = make(map[string]Student)
	//processor = sampleProcessor
)

type (
	// Session Holds info about a session
	Session map[string]interface{}

	// JSON Holds a JSON object
	JSON map[string]interface{}

	// Processor Alias for Process func
	//Processor func(session Session, message string) (string, error)

	Student struct {
		password string
		InnerDB  []Deadline
		//	schedule Schedule

	}
	Deadline struct {
		DeadlineType int
		Date         time.Time
		Slot         int
		CourseName   string
		WeekDay      string
		Venue        string
		Seat         string
	}
)

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	//tok, err := tokenFromFile(cacheFile)
	//if err != nil {
	tok := getTokenFromWeb(config)
	saveToken(cacheFile, tok)
	//}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calendar-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// writeJSON Writes the JSON equivilant for data into ResponseWriter w
func writeJSON(w http.ResponseWriter, data JSON) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handle Handles /
func handle(w http.ResponseWriter, r *http.Request) {
	body :=
		"<!DOCTYPE html><html><head><title>Chatbot</title></head><body><pre style=\"font-family: monospace;\">\n" +
			"Available Routes:\n\n" +
			"  GET  /welcome -> handleWelcome\n" +
			"  POST /chat    -> handleChat\n" +
			"  GET  /        -> handle        (current)\n" +
			"</pre></body></html>"
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintln(w, body)
}

func handleWelcome(w http.ResponseWriter, r *http.Request) {
	// Generate a UUID.
	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	uuid := hex.EncodeToString(hasher.Sum(nil))

	// Create a session for this UUID

	sessions[uuid] = Session{}

	// Write a JSON containg the welcome message and the generated UUID
	writeJSON(w, JSON{
		"uuid":    uuid,
		"message": WelcomeMessage,
	})
}

func processor(session Session, DeadlineType string, Datee string, Slott string, IOO string, Course string, venue string, interval string) (string, error) {
	out := ""
	var errorr error

	users, _ := session["username"].(string)
	u := DB[users].InnerDB
	if IOO == "in" {
		dead := Deadline{}

		switch DeadlineType {
		case "Quiz":
			dead.DeadlineType = 2

			dead.Date, errorr = time.Parse("2006-01-02", Datee)
			dead.Slot, errorr = strconv.Atoi(Slott)
			dead.CourseName = Course
			dead.Venue = venue
			u = append(u, dead)
			DB[users] = Student{DB[users].password, u}
			out = "Done"
			break
		case "Assignment":
			dead.DeadlineType = 3
			dead.Date, errorr = time.Parse("2006-01-02", Datee)
			dead.CourseName = Course
			u = append(u, dead)
			DB[users] = Student{DB[users].password, u}
			out = "Done"
			break
		case "Project":
			dead.DeadlineType = 4
			dead.Date, errorr = time.Parse("2006-01-02", Datee)
			dead.CourseName = Course
			u = append(u, dead)
			DB[users] = Student{DB[users].password, u}
			out = "Done"
			break
		default:
			out = "Wrong DeadlineType"
			errorr = nil
		}
	} else if IOO == "out" {
		switch DeadlineType {
		case "Schedule":
			for k, val := range u {
				if val.DeadlineType == 0 {
					val2 := u[k].WeekDay
					day := time.Now().Weekday().String()
					switch interval {
					case "Today":
						if strings.Compare(val2, day) == 0 {
							out += "Course :" + u[k].CourseName + string(u[k].Slot) + " Slot "
						}
						break
					case "This week":
						out += "Course :" + u[k].CourseName + string(u[k].Slot) + " Slot "
						break
					default:
						if strings.Compare(interval, val2) == 0 {
							out += "Course :" + u[k].CourseName + string(u[k].Slot) + " Slot "
						}
					}

				}
			}
			break
		case "Exam":
			for k, val := range u {
				if val.DeadlineType == 1 {
					val2 := u[k].Date
					day := time.Now()
					switch interval {
					case "Today":
						if val2.Equal(day) {
							out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02 15:04:05") + " Location:" + string(u[k].Venue) + " Seat:" + string(u[k].Seat)
						}
						break
					case "This week":
						day2 := day.Add(0000 - 00 - 07)
						if val2.Before(day2) && val2.After(day) {
							out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02 15:04:05") + " Location:" + string(u[k].Venue) + " Seat:" + string(u[k].Seat)
						}
						break
					default:

						out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02 15:04:05") + " Location:" + string(u[k].Venue) + " Seat:" + string(u[k].Seat)

					}

				}
			}

			break
		case "Quiz":
			for k, val := range u {
				if val.DeadlineType == 2 {
					val2 := u[k].Date
					day := time.Now().Format("2006-01-02")
					dayy, _ := time.Parse("2006-01-02", day)
					switch interval {
					case "Today":
						if val2.Equal(dayy) {
							out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02") + " Location:" + string(u[k].Venue) + " Slot:" + string(u[k].Slot)
						}
						break
					case "This week":
						day2 := dayy.AddDate(0, 0, 7)
						if val2.Before(day2) && val2.After(dayy) {
							out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02") + " Location:" + string(u[k].Venue) + " Slot:" + string(u[k].Slot)
						}
						break
					default:
						out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02") + " Location:" + string(u[k].Venue) + " Slot:" + string(u[k].Slot)
					}

				}
			}

			break
		case "Assignment":
			for k, val := range u {
				if val.DeadlineType == 3 {
					val2 := u[k].Date
					day := time.Now().Format("2006-01-02")
					dayy, _ := time.Parse("2006-01-02", day)
					switch interval {
					case "Today":
						if val2.Equal(dayy) {
							out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02 15:04:05")
						}
						break
					case "This week":
						day2 := dayy.AddDate(0, 0, 7)
						if val2.Before(day2) && val2.After(dayy) {
							out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02 15:04:05")
						}
						break

					default:
						out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02 15:04:05")
					}

				}
			}
			break
		case "Project":
			for k, val := range u {
				if val.DeadlineType == 4 {
					val2 := u[k].Date
					day := time.Now().Format("2006-01-02")
					dayy, _ := time.Parse("2006-01-02", day)
					switch interval {
					case "Today":
						if val2.Equal(dayy) {
							out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02 15:04:05")
						}
						break
					case "This week":
						day2 := dayy.AddDate(0, 0, 7)
						if val2.Before(day2) && val2.After(dayy) {
							out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02 15:04:05")
						}
						break

					default:
						out += " Course :" + u[k].CourseName + " Date:" + (u[k].Date).Format("2006-01-02 15:04:05")
					}

				}
			}

			break
		default:
			out = "Wrong DeadlineType"
			errorr = nil
		}
	} else {
		out = "Wrong direction"
		errorr = nil
	}
	return out, errorr
}

func calendarFunc(session Session, in string) (string, error) {
	out := ""
	var errorr error

	users, _ := session["username"].(string)
	u := DB[users].InnerDB

	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		return "Unable to read client secret file: ", err
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/calendar-go-quickstart.json
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return "Unable to parse client secret file to config: ", err
	}
	if strings.Compare(session["Phaser"].(string), "googlelink") == 0 {
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		out = "Go to the following link in your browser then type the authorization code:   " + authURL
		session["Phaser"] = "googletoken"
		return out, nil
	}
	cacheFile, err := tokenCacheFile()
	if err != nil {
		return "Unable to get path to cached credential file. ", err
	}

	if strings.Compare(session["Phaser"].(string), "googletoken") == 0 {
		session["Phaser"] = "3ayzeh"
	}
	tok, err := config.Exchange(oauth2.NoContext, in)
	if err != nil {
		return "Unable to retrieve token from web ", err
	}

	saveToken(cacheFile, tok)
	client := config.Client(ctx, tok)

	srv, err := calendar.New(client)
	if err != nil {
		return "Unable to retrieve calendar Client ", err
	}
	for _, val := range u {
		if val.DeadlineType != 0 {
			summary := ""
			switch val.DeadlineType {
			case 1:
				summary = "Exam"
				break
			case 2:
				summary = "Quiz"
				break
			case 3:
				summary = "Assignment"
				break
			case 4:
				summary = "Project"
				break
			default:
			}
			twoHours := time.Hour * -2
			exx := (val.Date).Add(twoHours)
			event := &calendar.Event{
				Summary:     summary,
				Location:    "German University In Cairo",
				Description: "You have " + summary + " in Course: " + val.CourseName + " in Seat: " + val.Seat + " at Location: " + val.Venue + " Slot: " + string(val.Slot),
				Start: &calendar.EventDateTime{

					DateTime: (exx).Format("2006-01-02T15:04:05Z07:00"),
					TimeZone: "Africa/Cairo",
				},

				End: &calendar.EventDateTime{
					DateTime: (val.Date).Format("2006-01-02T15:04:05Z07:00"),
					TimeZone: "Africa/Cairo",
				},

				//Recurrence: []string{"RRULE:FREQ=DAILY;COUNT=2"},
				/*Reminders: &calendar.EventReminders{
					//UseDefault: false,
					Overrides: []*calendar.EventReminder{
						{Minutes: 24 * 60},
						{Minutes: 60},
					},
				},*/
			}

			calendarId := "primary"
			event, err = srv.Events.Insert(calendarId, event).Do()
			if err != nil {
				return "Unable to create event. ", err
			}

		}
	}
	out = "Your deadlines were added to your google calendar successfully"
	return out, errorr
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	// Make sure only POST requests are handled
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}
	// Make sure a UUID exists in the Authorization header
	uuid := r.Header.Get("Authorization")
	if uuid == "" {
		http.Error(w, "Missing or empty Authorization header.", http.StatusUnauthorized)
		return
	}
	// Make sure a session exists for the extracted UUID
	session, sessionFound := sessions[uuid]
	if !sessionFound {
		http.Error(w, fmt.Sprintf("No session found for: %v.", uuid), http.StatusUnauthorized)
		return
	}
	_, userSession := session["Phaser"]
	if !userSession {
		session["Phaser"] = "username"
	}
	bdy := JSON{}
	if err := json.NewDecoder(r.Body).Decode(&bdy); err != nil {
		http.Error(w, fmt.Sprintf("Couldn't decode JSON: %v.", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	_, messageFound := bdy["message"]
	if !messageFound {
		http.Error(w, "Missing Message key in body.", http.StatusBadRequest)
		return
	}
	_, userSession2 := session["un"]
	if !userSession2 {
		session["un"] = ""
	}
	_, userSession3 := session["pw"]
	if !userSession3 {
		session["pw"] = ""
	}
	_, userSession4 := session["Date"]
	if !userSession4 {
		session["date"] = ""
	}
	_, userSession5 := session["Slot"]
	if !userSession5 {
		session["slot"] = ""
	}
	_, userSession6 := session["Course"]
	if !userSession6 {
		session["course"] = ""
	}
	_, userSession7 := session["Venue"]
	if !userSession7 {
		session["venue"] = ""
	}
	_, userSession8 := session["Interval"]
	if !userSession8 {
		session["interval"] = ""
	}
	_, userSession9 := session["direction"]
	if !userSession9 {
		session["direction"] = ""
	}
	_, userSession10 := session["DeadlineType"]
	if !userSession10 {
		session["DeadlineType"] = ""
	}

	if strings.Compare(session["Phaser"].(string), "username") == 0 {
		fmt.Println("Phase: " + session["Phaser"].(string))
		session["Phaser"] = "password"
		writeJSON(w, JSON{"message": "Please Enter Your GUC Username"})
		return
	}
	if strings.Compare(session["Phaser"].(string), "password") == 0 {
		fmt.Println("Phase: " + session["Phaser"].(string))
		session["un"] = bdy["message"].(string)
		session["Phaser"] = "login"
		writeJSON(w, JSON{"message": "Please Enter Your GUC Password"})
		return
	}
	if strings.Compare(session["Phaser"].(string), "login") == 0 {
		fmt.Println("Phase: " + session["Phaser"].(string))
		session["pw"] = bdy["message"].(string)
		mm := Login(session, session["un"].(string), session["pw"].(string))
		if strings.Compare(mm, "Wrong Username or Password") == 0 {
			session["Phaser"] = "username"
			writeJSON(w, JSON{"message": mm})
			return
		} else {
			session["Phaser"] = "3ayzeh"
			writeJSON(w, JSON{"message": mm})
			return
		}

	}
	if strings.Compare(session["Phaser"].(string), "3ayzeh") == 0 {
		fmt.Println("Phase: " + session["Phaser"].(string))
		session["Phaser"] = "direction"
		writeJSON(w, JSON{"message": "Do you want to enter a new deadline or ask for one?"})
		return
	}
	if strings.Compare(session["Phaser"].(string), "direction") == 0 {
		fmt.Println("Phase: " + session["Phaser"].(string))
		if strings.Compare(bdy["message"].(string), "enter one") == 0 {
			session["direction"] = "in"
			session["Phaser"] = "deadlinetype"
		} else if strings.Compare(bdy["message"].(string), "ask for one") == 0 {
			session["direction"] = "out"
			session["Phaser"] = "deadlinetype"
		}
		writeJSON(w, JSON{"message": "Enter your deadline type"})
		return //2017-11-05,3,Embedded,c7.02
	}
	if strings.Compare(session["Phaser"].(string), "deadlinetype") == 0 {
		fmt.Println("Phase: " + session["Phaser"].(string))
		if strings.Compare(bdy["message"].(string), "quiz") == 0 {
			session["DeadlineType"] = "Quiz"
			if strings.Compare(session["direction"].(string), "out") == 0 {
				session["Phaser"] = "interval"
				writeJSON(w, JSON{"message": "This week or today or all quizzes?"})
				return
			} else {
				session["Phaser"] = "7agtakquiz"
				writeJSON(w, JSON{"message": "please enter the following in this format, date:yyyy-mm-dd , quiz slot number(must be a number), coursename , quiz location with commas in between"})
				return
				//date2017-mm-dd--slot-coursename-venue
			}

		} else {
			if strings.Compare(bdy["message"].(string), "assignment") == 0 {
				session["DeadlineType"] = "Assignment"
				if strings.Compare(session["direction"].(string), "out") == 0 {
					session["Phaser"] = "interval"
					writeJSON(w, JSON{"message": "This week or today or all assignments?"})
					return
				} else {
					session["Phaser"] = "7agtakassignment"
					writeJSON(w, JSON{"message": "please enter the following in this format, date:yyyy-mm-dd, coursename"})
					return
					//date2017-mm-dd--slot-coursename-venue
				}

			} else {
				if strings.Compare(bdy["message"].(string), "project") == 0 {
					session["DeadlineType"] = "Project"
					if strings.Compare(session["direction"].(string), "out") == 0 {
						session["Phaser"] = "interval"
						writeJSON(w, JSON{"message": "This week or today or all projects?"})
						return
					} else {
						session["Phaser"] = "7agtakproject"
						writeJSON(w, JSON{"message": "please enter the following in this format, date:yyyy-mm-dd, coursename"})
						return
						//date2017-mm-dd--slot-coursename-venue
					}

				} else {
					if strings.Compare(bdy["message"].(string), "schedule") == 0 {
						session["DeadlineType"] = "Schedule"
						if strings.Compare(session["direction"].(string), "out") == 0 {
							session["Phaser"] = "interval"
							writeJSON(w, JSON{"message": "This week or today or Specific Day?"})
							return
						} else {
							session["Phaser"] = "3ayzeh"
							writeJSON(w, JSON{"message": "You cannot Post A schedule! It is automatically loaded from your system"})
							return
							//date2017-mm-dd--slot-coursename-venue
						}

					} else {
						if strings.Compare(bdy["message"].(string), "exam") == 0 {
							session["DeadlineType"] = "Exam"
							if strings.Compare(session["direction"].(string), "out") == 0 {
								session["Phaser"] = "interval"
								writeJSON(w, JSON{"message": "This week or today or all Exams?"})
								return
							} else {
								session["Phaser"] = "3ayzeh"
								writeJSON(w, JSON{"message": "You cannot Post An Exam! It is automatically loaded from your system"})
								return
								//date2017-mm-dd--slot-coursename-venue
							}

						} else {
							if strings.Compare(bdy["message"].(string), "all") == 0 {
								session["DeadlineType"] = "Exam"
								if strings.Compare(session["direction"].(string), "out") == 0 {
									session["Phaser"] = "googlelink"
									m, e := calendarFunc(session, bdy["message"].(string))
									writeJSON(w, JSON{"message": m, "error": e})
									return
								} else {
									session["Phaser"] = "3ayzeh"
									writeJSON(w, JSON{"message": "maynfa3shii"})
									return
									//date2017-mm-dd--slot-coursename-venue
								}
							}
						}
					}
				}
			}
		}
	}

	if strings.Compare(session["Phaser"].(string), "googletoken") == 0 {

		m, e := calendarFunc(session, bdy["message"].(string))
		writeJSON(w, JSON{"message": m, "error": e})
		return
	}
	//
	if strings.Compare(session["Phaser"].(string), "7agtakquiz") == 0 {
		fmt.Println("Phase: " + session["Phaser"].(string))
		s := strings.Split(bdy["message"].(string), ",")
		session["date"] = s[0]
		session["slot"] = s[1]
		session["course"] = s[2]
		session["venue"] = s[3]
		session["Phaser"] = "3ayzeh"
		fmt.Println(s[0])
		fmt.Println(session["date"].(string))
		fmt.Println(session["slot"].(string))
		fmt.Println(session["course"].(string))
		fmt.Println(session["venue"].(string))
	} else {
		if strings.Compare(session["Phaser"].(string), "7agtakassignment") == 0 {
			fmt.Println("Phase: " + session["Phaser"].(string))
			s := strings.Split(bdy["message"].(string), ",")
			session["date"] = s[0]

			session["course"] = s[1]

			session["Phaser"] = "3ayzeh"
			fmt.Println(s[0])
			fmt.Println(session["date"].(string))

			fmt.Println(session["course"].(string))

		} else {
			if strings.Compare(session["Phaser"].(string), "7agtakproject") == 0 {
				fmt.Println("Phase: " + session["Phaser"].(string))
				s := strings.Split(bdy["message"].(string), ",")
				session["date"] = s[0]

				session["course"] = s[1]

				session["Phaser"] = "3ayzeh"
				fmt.Println(s[0])
				fmt.Println(session["date"].(string))

				fmt.Println(session["course"].(string))

			}
		}
	}
	if strings.Compare(session["Phaser"].(string), "interval") == 0 {
		fmt.Println("Phase: " + session["Phaser"].(string))
		if strings.Compare(bdy["message"].(string), "this week") == 0 {
			session["interval"] = "This week"
		} else if strings.Compare(bdy["message"].(string), "today") == 0 {
			session["interval"] = "Today"
		} else if strings.Compare(bdy["message"].(string), "all") == 0 {
			session["interval"] = "all"
		} else {
			session["interval"] = bdy["message"].(string)
		}
		session["Phaser"] = "3ayzeh"
	}

	message, err := processor(session, session["DeadlineType"].(string), session["date"].(string), session["slot"].(string), session["direction"].(string), session["course"].(string), session["venue"].(string), session["interval"].(string))
	if err != nil {
		http.Error(w, err.Error(), 422 /* http.StatusUnprocessableEntity */)
		return
	}
	fmt.Println(DB)

	// Write a JSON containg the processed response
	writeJSON(w, JSON{
		"message": message,
		"Error":   err,
	})
}

func Login(session Session, username string, password string) string {
	reVal := "empty"
	Auth := factory.IsUserAuthorized(username, password)
	if Auth {
		_, userSession := session["username"]
		if !userSession {
			session["username"] = ""
		}
		session["username"] = username
		_, userFound := DB[username]
		if !userFound {
			if schedules, err := factory.GetUserSchedule(username, password); err != nil {
				reVal = "Unauthorized GUC access"
				fmt.Println(reVal)
			} else {
				dead := []Deadline{}
				for k := range schedules {
					gdida := Deadline{}
					gdida.CourseName = schedules[k].Course
					gdida.DeadlineType = 0
					gdida.Slot = schedules[k].Slot
					gdida.WeekDay = schedules[k].Weekday
					dead = append(dead, gdida)
				}
				if exams, err := factory.GetUserExams(username, password); err != nil {
					reVal = "Unauthorized GUC access"
					fmt.Println(reVal)
				} else {
					for k := range exams {
						TempExam := Deadline{}
						TempExam.CourseName = exams[k].Course
						TempExam.DeadlineType = 1
						TempExam.Date = exams[k].DateTime
						TempExam.Venue = exams[k].Venue
						TempExam.Seat = exams[k].Seat
						dead = append(dead, TempExam)
					}
				}
				student := &Student{password, dead}
				DB[username] = *student
				//json.NewEncoder(w).Encode(DB[username])
				reVal = "Login Successful, Your schedule has been added"
				fmt.Println(reVal)
			}
		} else {
			reVal = "Login Successful"
			fmt.Println(reVal)
		}
	} else {
		reVal = "Wrong Username or Password"
	}
	return reVal
}

func sendUnauthorizedJSONResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	util.SendJSONResponse(w, factory.ResponseAPI{Error: err.Error(), Data: nil})
}

func sendDataJSONResponse(w http.ResponseWriter, data interface{}) {
	util.SendJSONResponse(w, factory.ResponseAPI{Error: nil, Data: data})
}
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/welcome", handleWelcome)
	mux.HandleFunc("/chat", handleChat)
	mux.HandleFunc("/", handle)

	// Start the server
	log.Fatalln(http.ListenAndServe(":3011", cors.CORS(mux)))

}
