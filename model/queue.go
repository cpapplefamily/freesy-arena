package model

import "sort"
import "time"

type QueueItem struct {
	TeamId          int `db:"id,manual"`
	QueueTime	   	time.Time
	Nickname        string
	RedPreference   bool
	BluePreference  bool
	Red1Preference  bool
	Red2Preference  bool
	Red3Preference  bool
	Blue1Preference bool
	Blue2Preference bool
	Blue3Preference bool
}

func (database *Database) CreateQueueItem(queue_item *QueueItem) error {
	return database.queueTable.create(queue_item)
}

func (database *Database) GetQueueItemById(id int) (*QueueItem, error) {
	return database.queueTable.getById(id)
}

func (database *Database) UpdateQueueItem(queue_item *QueueItem) error {
	return database.queueTable.update(queue_item)
}

func (database *Database) DeleteQueueItem(id int) error {
	return database.queueTable.delete(id)
}

func (database *Database) TruncateQueueItems() error {
	return database.queueTable.truncate()
}

func (database *Database) GetAllQueueItems() ([]QueueItem, error) {
	queue_item, err := database.queueTable.getAll()
	if err != nil {
		return nil, err
	}
	sort.Slice(queue_item, func(i, j int) bool {
		return queue_item[i].QueueTime.Before(queue_item[j].QueueTime)
	})
	return queue_item, nil
}
