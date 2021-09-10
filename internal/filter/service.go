package filter

import (
	"github.com/rs/zerolog/log"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/internal/indexer"
)

type Service interface {
	//FindFilter(announce domain.Announce) (*domain.Filter, error)

	FindByID(filterID int) (*domain.Filter, error)
	FindByIndexerIdentifier(announce domain.Announce) (*domain.Filter, error)
	ListFilters() ([]domain.Filter, error)
	Store(filter domain.Filter) (*domain.Filter, error)
	Update(filter domain.Filter) (*domain.Filter, error)
	Delete(filterID int) error
}

type service struct {
	repo       domain.FilterRepo
	actionRepo domain.ActionRepo
	indexerSvc indexer.Service
}

func NewService(repo domain.FilterRepo, actionRepo domain.ActionRepo, indexerSvc indexer.Service) Service {
	return &service{
		repo:       repo,
		actionRepo: actionRepo,
		indexerSvc: indexerSvc,
	}
}

func (s *service) ListFilters() ([]domain.Filter, error) {
	// get filters
	filters, err := s.repo.ListFilters()
	if err != nil {
		return nil, err
	}

	var ret []domain.Filter

	for _, filter := range filters {
		indexers, err := s.indexerSvc.FindByFilterID(filter.ID)
		if err != nil {
			return nil, err
		}
		filter.Indexers = indexers

		ret = append(ret, filter)
	}

	return ret, nil
}

func (s *service) FindByID(filterID int) (*domain.Filter, error) {
	// find filter
	filter, err := s.repo.FindByID(filterID)
	if err != nil {
		return nil, err
	}

	// find actions and attach
	//actions, err := s.actionRepo.FindFilterActions(filter.ID)
	actions, err := s.actionRepo.FindByFilterID(filter.ID)
	if err != nil {
		log.Error().Msgf("could not find filter actions: %+v", &filter.ID)
	}
	filter.Actions = actions

	// find indexers and attach
	indexers, err := s.indexerSvc.FindByFilterID(filter.ID)
	if err != nil {
		log.Error().Err(err).Msgf("could not find indexers for filter: %+v", &filter.Name)
		return nil, err
	}
	filter.Indexers = indexers

	//log.Debug().Msgf("found filter: %+v", filter)

	return filter, nil
}

func (s *service) FindByIndexerIdentifier(announce domain.Announce) (*domain.Filter, error) {
	// get filter for tracker
	filters, err := s.repo.FindByIndexerIdentifier(announce.Site)
	if err != nil {
		log.Error().Err(err).Msgf("could not find filters for indexer: %v", announce.Site)
		return nil, err
	}

	// match against announce/releaseInfo
	for _, filter := range filters {
		// if match, return the filter
		matchedFilter := s.checkFilter(filter, announce)
		if matchedFilter {
			log.Trace().Msgf("found matching filter: %+v", &filter)
			log.Debug().Msgf("found matching filter: %v", &filter.Name)

			// find actions and attach
			actions, err := s.actionRepo.FindByFilterID(filter.ID)
			if err != nil {
				log.Error().Err(err).Msgf("could not find filter actions: %+v", &filter.ID)
				return nil, err
			}

			// if no actions found, check next filter
			if actions == nil {
				continue
			}

			filter.Actions = actions

			return &filter, nil
		}
	}

	// if no match, return nil
	return nil, nil
}

//func (s *service) FindFilter(announce domain.Announce) (*domain.Filter, error) {
//	// get filter for tracker
//	filters, err := s.repo.FindFiltersForSite(announce.Site)
//	if err != nil {
//		return nil, err
//	}
//
//	// match against announce/releaseInfo
//	for _, filter := range filters {
//		// if match, return the filter
//		matchedFilter := s.checkFilter(filter, announce)
//		if matchedFilter {
//
//			log.Debug().Msgf("found filter: %+v", &filter)
//
//			// find actions and attach
//			actions, err := s.actionRepo.FindByFilterID(filter.ID)
//			if err != nil {
//				log.Error().Msgf("could not find filter actions: %+v", &filter.ID)
//			}
//			filter.Actions = actions
//
//			return &filter, nil
//		}
//	}
//
//	// if no match, return nil
//	return nil, nil
//}

func (s *service) Store(filter domain.Filter) (*domain.Filter, error) {
	// validate data

	// store
	f, err := s.repo.Store(filter)
	if err != nil {
		log.Error().Err(err).Msgf("could not store filter: %v", filter)
		return nil, err
	}

	return f, nil
}

func (s *service) Update(filter domain.Filter) (*domain.Filter, error) {
	// validate data

	// store
	f, err := s.repo.Update(filter)
	if err != nil {
		log.Error().Err(err).Msgf("could not update filter: %v", filter.Name)
		return nil, err
	}

	// take care of connected indexers
	if err = s.repo.DeleteIndexerConnections(f.ID); err != nil {
		log.Error().Err(err).Msgf("could not delete filter indexer connections: %v", filter.Name)
		return nil, err
	}

	for _, i := range filter.Indexers {
		if err = s.repo.StoreIndexerConnection(f.ID, int(i.ID)); err != nil {
			log.Error().Err(err).Msgf("could not store filter indexer connections: %v", filter.Name)
			return nil, err
		}
	}

	return f, nil
}

func (s *service) Delete(filterID int) error {
	if filterID == 0 {
		return nil
	}

	// delete
	if err := s.repo.Delete(filterID); err != nil {
		log.Error().Err(err).Msgf("could not delete filter: %v", filterID)
		return err
	}

	return nil
}
