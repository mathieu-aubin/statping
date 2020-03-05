package services

import (
	"errors"
	"fmt"
	"github.com/hunterlong/statping/database"
	"github.com/hunterlong/statping/utils"
)

var log = utils.Log

func DB() database.Database {
	return database.DB().Model(&Service{})
}

func Find(id int64) (*Service, error) {
	srv := allServices[id]
	if srv == nil {
		return nil, errors.New("service not found")
	}
	return srv, nil
}

func All() []*Service {
	var services []*Service
	DB().Find(&services)
	return services
}

func (s *Service) Create() error {
	err := DB().Create(&s)
	if err.Error() != nil {
		log.Errorln(fmt.Sprintf("Failed to create service %v #%v: %v", s.Name, s.Id, err))
		return err.Error()
	}
	allServices[s.Id] = s

	go ServiceCheckQueue(allServices[s.Id], true)
	reorderServices()
	//notifications.OnNewService(s)

	return nil
}

func (s *Service) Update() error {
	db := DB().Update(&s)

	allServices[s.Id] = s

	if !s.AllowNotifications.Bool {
		//for _, n := range CoreApp.Notifications {
		//	notif := n.(notifier.Notifier).Select()
		//	notif.ResetUniqueQueue(fmt.Sprintf("service_%v", s.Id))
		//}
	}
	s.Close()
	s.SleepDuration = s.Duration()
	go ServiceCheckQueue(allServices[s.Id], true)

	reorderServices()
	//notifier.OnUpdatedService(s.Service)

	return db.Error()
}

func (s *Service) Delete() error {
	db := database.DB().Delete(&s)

	s.Close()
	delete(allServices, s.Id)
	reorderServices()
	//notifier.OnDeletedService(s.Service)

	return db.Error()
}

func (s *Service) DeleteFailures() error {
	query := database.DB().Exec(`DELETE FROM failures WHERE service = ?`, s.Id)
	return query.Error()
}