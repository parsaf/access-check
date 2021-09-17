package controller

import (
	"context"
	"time"
	"fmt"

	"github.com/pkg/errors"
)

const InputTimestampFormat = "1/2/2006 15:04:05"

type UserAccess struct {
	User string `json:"user"`
	Grantor string `json:"grantor"`
	CreateTimestamp string `json:"createTimestamp"`
	EndTimeStamp string `json:"endTimeStamp"`

}

func (c Controller) WhoHadAccess(
	ctx context.Context,
	role string,
	timestamp string,
)([]UserAccess, error)  {
	accessTime, err := time.Parse(InputTimestampFormat, timestamp)
	fmt.Println("time:", accessTime, "role: ", role)
	if err != nil {
		errors.Wrap(err, "Bad timestamp input")
	}
	accessIntervals, err := c.store.GetByTimeAndRole(
		ctx,
		accessTime,
		role,
	)
	fmt.Println("Intervals:", accessIntervals)
	if err != nil {
		return nil, err
	}
	userAccesses := make([]UserAccess, len(accessIntervals))
	for i, interval := range accessIntervals {
		userAccesses[i] = UserAccess{
			User:            interval.User,
			Grantor:         interval.Grantor,
			CreateTimestamp: interval.StartTime.Format(InputTimestampFormat),
			EndTimeStamp:    interval.EndTime.Format(InputTimestampFormat),
		}
	}
	return userAccesses, nil
}