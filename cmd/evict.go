package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/irgendwr/go-ts3"
	"github.com/pkg/errors"
)

type onlineClient struct {
	ID                 int    `ms:"cid"`
	CLID               int    `ms:"clid"`
	UniqueIdentifier   string `ms:"client_unique_identifier"`
	ConnectionClientIP string `ms:"connection_client_ip"`
	DatabaseID         int    `ms:"client_database_id"`
	Nickname           string `ms:"client_nickname"`
	Type               int    `ms:"client_type"`
	Away               bool   `ms:"client_away"`
	AwayMessage        string `ms:"client_away_message"`
	LastConnected      int64  `ms:"client_lastconnected"` // FIXME: time.Time
	IdleTime           int64  `ms:"client_idle_time"`     // FIXME: time.Time
	Servergroups       string `ms:"client_servergroups"`
}

type violationEntry struct {
	Count    int
	Violator violator
}

type violator struct {
	UID  string
	Nick string
}

func evict(cfg config) error {
	var violationsFile *os.File
	violators := []violationEntry{}

	if cfg.Violators != "" {
		f, err := os.OpenFile(cfg.Violators, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return errors.Wrapf(err, "unable to open file %q", cfg.Violators)
		}
		violationsFile = f

		r := csv.NewReader(violationsFile)
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			count, err := strconv.Atoi(record[0])
			if err != nil {
				return err
			}
			if len(record) < 2 {
				return errors.New("invalid record length (at least two fields required)")
			}
			nick := ""
			if len(record) >= 3 {
				nick = record[2]
			}

			violators = append(violators, violationEntry{
				Count: count,
				Violator: violator{
					UID:  record[1],
					Nick: nick,
				},
			})
		}
	}

	for _, s := range cfg.Servers {
		if newViolators, err := s.evict(cfg, violators); err != nil {
			log.Printf("Error: %s\n", err)
		} else {
			for _, newViolation := range newViolators {
				match := false
				for i := range violators {
					if newViolation.UID == violators[i].Violator.UID {
						violators[i].Count++
						violators[i].Violator.Nick = newViolation.Nick
						match = true
						break
					}
				}
				if !match {
					violators = append(violators, violationEntry{
						Count:    1,
						Violator: newViolation,
					})
				}
			}
		}
	}

	if violationsFile != nil {
		if _, err := violationsFile.Seek(0, 0); err != nil {
			return err
		}

		violationsWriter := csv.NewWriter(violationsFile)
		for _, violator := range violators {
			if err := violationsWriter.Write([]string{strconv.Itoa(violator.Count), violator.Violator.UID, violator.Violator.Nick}); err != nil {
				return errors.Wrap(err, "error writing entry")
			}
		}

		violationsWriter.Flush()
		if err := violationsWriter.Error(); err != nil {
			return err
		}
	}

	return nil
}

func (s server) evict(cfg config, violations []violationEntry) ([]violator, error) {
	newViolators := []violator{}
	s.fillDefaults(cfg)

	addr := fmt.Sprintf("%s:%d", s.IP, s.QueryPort)
	log.Printf("Checking %s...\n", addr)
	c, err := ts3.NewClient(addr)
	if err != nil {
		return newViolators, err
	}
	defer c.Close()

	if err := c.Login(s.Username, s.Password); err != nil {
		return newViolators, err
	}

	for _, port := range s.Ports {
		var wg sync.WaitGroup
		var mutex sync.Mutex

		if err := c.UsePort(port); err != nil {
			log.Printf("Error: Invalid port '%d' on host '%s'\n", port, s.IP)
			continue
		}

		var groups []*ts3.Group
		if groups, err = c.Server.GroupList(); err != nil {
			return newViolators, err
		}
		var ignoreGroups []*ts3.Group
		for _, group := range groups {
			for _, name := range cfg.IgnoreGroupNames {
				if strings.ToUpper(group.Name) == strings.ToUpper(name) {
					ignoreGroups = append(ignoreGroups, group)
				}
			}
		}

		var clients []*onlineClient
		if _, err := c.ExecCmd(ts3.NewCmd("clientlist").WithOptions("-uid", "-times", "-groups", "-info").WithResponse(&clients)); err != nil {
			return newViolators, err
		}

		for _, client := range clients {
			var clientViolations *violationEntry
			overLimit := false
			for _, violation := range violations {
				if client.UniqueIdentifier == violation.Violator.UID {
					clientViolations = &violation
					if cfg.Action == "kick or ban" {
						if violation.Count >= cfg.Kicklimit {
							overLimit = true
						}
					}
					break
				}
			}

			// FIXME: time.Since(client.LastConnected)
			duration := time.Since(time.Unix(client.LastConnected, 0))
			if duration >= time.Duration(cfg.Timelimit)*time.Minute || overLimit {
				// ignore query clients
				if client.Type == 1 {
					//log.Printf("Ignoring query client: %s\n", client.Nickname)
					continue
				}
				// ignore clients with given groups
				if hasGroup(client, ignoreGroups) {
					//log.Printf("Ignoring due to group: %s\n", client.Nickname)
					continue
				}

				newViolators = append(newViolators, violator{
					UID:  client.UniqueIdentifier,
					Nick: client.Nickname,
				})

				log.Printf("Messaging %s...\n", client.Nickname)
				mutex.Lock()
				if _, err := c.ExecCmd(ts3.NewCmd("sendtextmessage").WithArgs(ts3.NewArg("targetmode", 1), ts3.NewArg("target", client.CLID), ts3.NewArg("msg", cfg.Message))); err != nil {
					log.Printf("Error: Unable to send message: %s\n", err)
				}
				mutex.Unlock()

				action := cfg.Action
				if action == "kick or ban" {
					if clientViolations != nil && clientViolations.Count >= cfg.Kicklimit {
						action = "ban"
					} else {
						action = "kick"
					}
				}

				wg.Add(1)
				go func(client *onlineClient, action string) {
					defer wg.Done()

					if cfg.Delay > 0 {
						time.Sleep(time.Duration(cfg.Delay) * time.Second)
					}

					log.Printf("Evicting %s | %s | %s ...\n", client.Nickname, client.UniqueIdentifier, client.ConnectionClientIP)

					switch action {
					case "none":
						break
					case "ban":
						mutex.Lock()
						args := []ts3.CmdArg{
							ts3.NewArg("clid", client.CLID),
						}
						if cfg.BanMessage != "" {
							args = append(args, ts3.NewArg("banreason", cfg.BanMessage))
						}
						if cfg.BanDuration > 0 {
							args = append(args, ts3.NewArg("time", cfg.BanDuration))
						}
						if _, err := c.ExecCmd(ts3.NewCmd("banclient").WithArgs(args...)); err != nil {
							log.Printf("Error: Unable to ban %s: %s\n", client.Nickname, err)
						}
						mutex.Unlock()
					default:
						fallthrough
					case "kick":
						mutex.Lock()
						args := []ts3.CmdArg{
							ts3.NewArg("clid", client.CLID),
							ts3.NewArg("reasonid", 5 /* server kick*/),
						}
						if cfg.KickMessage != "" {
							args = append(args, ts3.NewArg("reasonmsg", cfg.KickMessage))
						}
						if _, err := c.ExecCmd(ts3.NewCmd("clientkick").WithArgs(args...)); err != nil {
							log.Printf("Error: Unable to kick %s: %s\n", client.Nickname, err)
						}
						mutex.Unlock()
					}
				}(client, action)
			}
		}
		wg.Wait()
	}

	return newViolators, nil
}

func hasGroup(c *onlineClient, groups []*ts3.Group) bool {
	for _, id := range strings.Split(c.Servergroups, ",") {
		for _, group := range groups {
			if id == strconv.Itoa(group.ID) {
				return true
			}
		}
	}
	return false
}

func (s *server) fillDefaults(c config) {
	if s.Username == "" {
		s.Username = c.DefaultUsername
	}
	if s.Password == "" {
		s.Password = c.DefaultPassword
	}
	if len(s.Ports) == 0 {
		s.Ports = c.DefaultPorts
	}
	if s.QueryPort == 0 {
		s.QueryPort = c.DefaultQueryPort
	}
}
