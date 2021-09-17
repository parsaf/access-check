package storage

import (
	"context"
	"time"

	"github.com/parsaf/access-check/model"
)

type AccessIntervalStorage interface {
	GetByTimeAndRole(ctx context.Context, accessTime time.Time, role string) ([]model.AccessInterval, error)
	LoadData(ctx context.Context, userToEventLogsMap map[string][]model.EventLog) error
}


type inMemoryStorage struct {
	// [role => [day => AccessLog]]
	roleToAccessDaysMap map[string]map[time.Time][]model.AccessInterval
}

var _ AccessIntervalStorage = &inMemoryStorage{}

func (s *inMemoryStorage) GetByTimeAndRole(ctx context.Context, accessTime time.Time, role string) ([]model.AccessInterval, error) {
	accessDaysMap, ok := s.roleToAccessDaysMap[role]
	
	if !ok {
		return nil, nil
	}

	// fetch the accessIntervals on the day of the requested time
	accessIntervalsOfDay := accessDaysMap[accessTime.Truncate(24 * time.Hour)]
	
	// aggregate intervals that contain the requested time
	var intervals []model.AccessInterval
	for _, interval := range accessIntervalsOfDay {
		// the inclusive start and non inclusive end of the interval
		start, end := interval.StartTime, interval.EndTime
		if accessTime.Before(start) || accessTime.After(end) || accessTime.Equal(end) {
			continue
		}
		intervals = append(intervals,interval)
	}
	return intervals, nil
}

func (s *inMemoryStorage) LoadData(ctx context.Context, userToEventLogsMap map[string][]model.EventLog) error {
	// AccessIntervals for current user
	var intervals []model.AccessInterval

	for user, logs := range userToEventLogsMap {
		for i, log := range logs {
			// Assumption: logs won't contain a user's role changing from A to A, only A to B
			
			// set endTime of previous interval
			if i > 0 {
				size := len(intervals)
				intervals[size-1].EndTime = log.Time
			}

			// new interval starting from current log
			intervals = append(intervals, model.AccessInterval{
				User:           user,
				Grantor:        log.Grantor,
				Role:           log.Role,
				StartTime: 		log.Time,
			})
		}

		// Assumption: timestamps won't go into the future
		// set endTime for the last interval to be now
		size := len(intervals)
		intervals[size-1].EndTime = time.Now()
	}

	// index intervals by role and day
	for _, interval := range intervals {
		role := interval.Role
		days := getDaysBetween(interval.StartTime, interval.EndTime)

		if _, ok := s.roleToAccessDaysMap[role]; !ok {
			s.roleToAccessDaysMap[role] = make(map[time.Time][]model.AccessInterval)
		}

		// Add interval to all days that it spans. This makes it easy to find if a 
		// Assumption: many roles are granted over intervals that span days. If more granularity
		// needed, swap for hours (memory permitting)
		for _, day := range days {
			intervalsOfDay := s.roleToAccessDaysMap[role][day]
			s.roleToAccessDaysMap[role][day] = append(intervalsOfDay, interval)
		}
	}


	return nil
}

func getDaysBetween(
	startTime time.Time,
	endTime time.Time,
) []time.Time {
	startDate := startTime.Truncate(24 * time.Hour)

	var days []time.Time
	for startDate.Before(endTime) {
		days = append(days, startDate)
		startDate = startDate.Add(24 * time.Hour)
	}
	return days
}

// NewInMemoryStorage returns a new store initialized with given logs
func NewInMemoryStorage(
	ctx context.Context,
	userToEventLogsMap map[string][]model.EventLog,
) (*inMemoryStorage, error) {
	store := &inMemoryStorage{
		roleToAccessDaysMap: make(map[string]map[time.Time][]model.AccessInterval),
	}
	err := store.LoadData(ctx, userToEventLogsMap)
	if err != nil {
		return nil, err
	}
	return store, nil
}


