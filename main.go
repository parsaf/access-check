package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parsaf/access-check/controller"
	"github.com/parsaf/access-check/model"
	"github.com/parsaf/access-check/storage"
	"github.com/pkg/errors"
)

const (
	inputFilePath = "./data/input.csv"
	delimiter     = '|'
)

func main() {
	// set up server
	userToEventLogsMap, err := readInputData()
	if err != nil {
		panic(err)
	}
	store, err := storage.NewInMemoryStorage(context.Background(), userToEventLogsMap)
	if err != nil {
		panic(err)
	}
	ctrl := controller.New(store)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	// e.g. /who?role=admin&timestamp=12/27/2020%209:30:00
	r.GET("/who", func(c *gin.Context) {
		role := c.Query("role")
		timestamp := c.Query("timestamp")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := ctrl.WhoHadAccess(c, role, timestamp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userAcesses": response})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

// readInputData reads in the logs from the input file and puts them in a map [user => list of EventLogs]
func readInputData() (map[string][]model.EventLog, error) {
	file, err := os.Open(inputFilePath)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = delimiter

	userToEventLogsMap := make(map[string][]model.EventLog)
	i := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// Ignore header
		if i == 0 {
			i++
			continue
		}
		i++

		timestamp := record[0]
		eventTime, err := time.Parse(controller.InputTimestampFormat, timestamp)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		grantor := record[1]
		user := record[2]
		role := record[3]
		eventLog := model.EventLog{
			User:    user,
			Grantor: grantor,
			Role:    role,
			Time:    eventTime,
		}

		eventLogs := userToEventLogsMap[user]
		userToEventLogsMap[user] = append(eventLogs, eventLog)
	}
	return userToEventLogsMap, nil
}
