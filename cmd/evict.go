package cmd

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/irgendwr/go-ts3"
)

type onlineClient struct {
	ID            int    `ms:"cid"`
	CLID          int    `ms:"clid"`
	DatabaseID    int    `ms:"client_database_id"`
	Nickname      string `ms:"client_nickname"`
	Type          int    `ms:"client_type"`
	Away          bool   `ms:"client_away"`
	AwayMessage   string `ms:"client_away_message"`
	LastConnected int64  `ms:"client_lastconnected"`
	IdleTime      int64  `ms:"client_idle_time"`
	Servergroups  []int  `ms:"client_servergroups"`
}

func evict(cfg config) error {
	for _, s := range cfg.Servers {
		if err := s.evict(cfg); err != nil {
			return err
		}
	}

	return nil
}

func (s server) evict(cfg config) error {
	s.fillDefaults(cfg)

	c, err := ts3.NewClient(fmt.Sprintf("%s:%d", s.IP, s.QueryPort))
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.Login(s.Username, s.Password); err != nil {
		return err
	}

	for _, port := range s.Ports {
		var wg sync.WaitGroup
		var mutex sync.Mutex

		if err := c.UsePort(port); err != nil {
			log.Printf("Invalid port '%d' on host '%s'\n", port, s.IP)
			continue
		}

		var groups []*ts3.Group
		if groups, err = c.Server.GroupList(); err != nil {
			return err
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
		if _, err := c.ExecCmd(ts3.NewCmd("clientlist").WithOptions("-times", "-groups").WithResponse(&clients)); err != nil {
			return err
		}

		for _, client := range clients {
			duration := time.Since(time.Unix(client.LastConnected, 0))
			if duration >= time.Duration(cfg.Timelimit)*time.Minute {
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

				log.Printf("messaging client %s...\n", client.Nickname)
				mutex.Lock()
				if _, err := c.ExecCmd(ts3.NewCmd("sendtextmessage").WithArgs(ts3.NewArg("targetmode", 1), ts3.NewArg("target", client.CLID), ts3.NewArg("msg", cfg.Message))); err != nil {
					log.Printf("Unable to send text message: %s\n", err)
				}
				mutex.Unlock()

				wg.Add(1)
				go func(client *onlineClient) {
					defer wg.Done()

					time.Sleep(time.Duration(cfg.Delay) * time.Second)

					log.Printf("evicting client %s...\n", client.Nickname)

					switch cfg.Action {
					case "none":
						break
					case "ban":
						mutex.Lock()
						if _, err := c.ExecCmd(ts3.NewCmd("banclient").WithArgs(ts3.NewArg("clid", client.CLID))); err != nil {
							log.Printf("Unable to ban %s: %s\n", client.Nickname, err)
						}
						mutex.Unlock()
					default:
					case "kick":
						mutex.Lock()
						if _, err := c.ExecCmd(ts3.NewCmd("clientkick").WithArgs(ts3.NewArg("clid", client.CLID), ts3.NewArg("reasonid", 5 /* server kick*/))); err != nil {
							log.Printf("Unable to kick %s: %s\n", client.Nickname, err)
						}
						mutex.Unlock()
					}
				}(client)
			}
		}
		wg.Wait()
	}

	return nil
}

func hasGroup(c *onlineClient, groups []*ts3.Group) bool {
	for _, id := range c.Servergroups {
		for _, group := range groups {
			if id == group.ID {
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