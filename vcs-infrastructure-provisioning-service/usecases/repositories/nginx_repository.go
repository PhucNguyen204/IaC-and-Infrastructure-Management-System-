package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

type INginxRepository interface {
	Create(instance *entities.NginxInstance) error
	GetByID(id string) (*entities.NginxInstance, error)
	FindByID(id string) (*entities.NginxInstance, error)
	FindByInfrastructureID(infraID string) (*entities.NginxInstance, error)
	Update(instance *entities.NginxInstance) error
	Delete(id string) error

	CreateDomain(domain *entities.NginxDomain) error
	DeleteDomain(nginxID, domain string) error
	ListDomains(nginxID string) ([]entities.NginxDomain, error)
	CreateRoute(route *entities.NginxRoute) error
	GetRoute(routeID string) (*entities.NginxRoute, error)
	UpdateRoute(route *entities.NginxRoute) error
	DeleteRoute(routeID string) error
	ListRoutes(nginxID string) ([]entities.NginxRoute, error)
	CreateOrUpdateCertificate(cert *entities.NginxCertificate) error
	GetCertificate(nginxID string) (*entities.NginxCertificate, error)
	CreateOrUpdateUpstream(upstream *entities.NginxUpstream) error
	DeleteUpstreamBackends(upstreamID string) error
	CreateUpstreamBackend(backend *entities.NginxUpstreamBackend) error
	ListUpstreams(nginxID string) ([]entities.NginxUpstream, error)
	CreateOrUpdateSecurity(security *entities.NginxSecurity) error
	GetSecurity(nginxID string) (*entities.NginxSecurity, error)
	DeleteSecurity(nginxID string) error
}

type nginxRepository struct {
	db *gorm.DB
}

func NewNginxRepository(db *gorm.DB) INginxRepository {
	return &nginxRepository{db: db}
}

func (r *nginxRepository) Create(instance *entities.NginxInstance) error {
	return r.db.Create(instance).Error
}

func (r *nginxRepository) FindByID(id string) (*entities.NginxInstance, error) {
	var instance entities.NginxInstance
	if err := r.db.Preload("Infrastructure").Where("id = ?", id).First(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *nginxRepository) FindByInfrastructureID(infraID string) (*entities.NginxInstance, error) {
	var instance entities.NginxInstance
	if err := r.db.Preload("Infrastructure").Where("infrastructure_id = ?", infraID).First(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *nginxRepository) Update(instance *entities.NginxInstance) error {
	return r.db.Save(instance).Error
}

func (r *nginxRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.NginxInstance{}).Error
}

func (r *nginxRepository) GetByID(id string) (*entities.NginxInstance, error) {
	return r.FindByID(id)
}

func (r *nginxRepository) CreateDomain(domain *entities.NginxDomain) error {
	return r.db.Create(domain).Error
}

func (r *nginxRepository) DeleteDomain(nginxID, domain string) error {
	return r.db.Where("nginx_id = ? AND domain = ?", nginxID, domain).Delete(&entities.NginxDomain{}).Error
}

func (r *nginxRepository) ListDomains(nginxID string) ([]entities.NginxDomain, error) {
	var domains []entities.NginxDomain
	if err := r.db.Where("nginx_id = ?", nginxID).Find(&domains).Error; err != nil {
		return nil, err
	}
	return domains, nil
}

func (r *nginxRepository) CreateRoute(route *entities.NginxRoute) error {
	return r.db.Create(route).Error
}

func (r *nginxRepository) GetRoute(routeID string) (*entities.NginxRoute, error) {
	var route entities.NginxRoute
	if err := r.db.Where("id = ?", routeID).First(&route).Error; err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *nginxRepository) UpdateRoute(route *entities.NginxRoute) error {
	return r.db.Save(route).Error
}

func (r *nginxRepository) DeleteRoute(routeID string) error {
	return r.db.Where("id = ?", routeID).Delete(&entities.NginxRoute{}).Error
}

func (r *nginxRepository) ListRoutes(nginxID string) ([]entities.NginxRoute, error) {
	var routes []entities.NginxRoute
	if err := r.db.Where("nginx_id = ?", nginxID).Order("priority DESC").Find(&routes).Error; err != nil {
		return nil, err
	}
	return routes, nil
}

func (r *nginxRepository) CreateOrUpdateCertificate(cert *entities.NginxCertificate) error {
	var existing entities.NginxCertificate
	if err := r.db.Where("nginx_id = ?", cert.NginxID).First(&existing).Error; err == nil {
		cert.ID = existing.ID
		return r.db.Save(cert).Error
	}
	return r.db.Create(cert).Error
}

func (r *nginxRepository) GetCertificate(nginxID string) (*entities.NginxCertificate, error) {
	var cert entities.NginxCertificate
	if err := r.db.Where("nginx_id = ?", nginxID).First(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

func (r *nginxRepository) CreateOrUpdateUpstream(upstream *entities.NginxUpstream) error {
	var existing entities.NginxUpstream
	if err := r.db.Where("nginx_id = ? AND name = ?", upstream.NginxID, upstream.Name).First(&existing).Error; err == nil {
		upstream.ID = existing.ID
		return r.db.Save(upstream).Error
	}
	return r.db.Create(upstream).Error
}

func (r *nginxRepository) DeleteUpstreamBackends(upstreamID string) error {
	return r.db.Where("upstream_id = ?", upstreamID).Delete(&entities.NginxUpstreamBackend{}).Error
}

func (r *nginxRepository) CreateUpstreamBackend(backend *entities.NginxUpstreamBackend) error {
	return r.db.Create(backend).Error
}

func (r *nginxRepository) ListUpstreams(nginxID string) ([]entities.NginxUpstream, error) {
	var upstreams []entities.NginxUpstream
	if err := r.db.Preload("Backends").Where("nginx_id = ?", nginxID).Find(&upstreams).Error; err != nil {
		return nil, err
	}
	return upstreams, nil
}

func (r *nginxRepository) CreateOrUpdateSecurity(security *entities.NginxSecurity) error {
	var existing entities.NginxSecurity
	if err := r.db.Where("nginx_id = ?", security.NginxID).First(&existing).Error; err == nil {
		security.ID = existing.ID
		return r.db.Save(security).Error
	}
	return r.db.Create(security).Error
}

func (r *nginxRepository) GetSecurity(nginxID string) (*entities.NginxSecurity, error) {
	var security entities.NginxSecurity
	if err := r.db.Where("nginx_id = ?", nginxID).First(&security).Error; err != nil {
		return nil, err
	}
	return &security, nil
}

func (r *nginxRepository) DeleteSecurity(nginxID string) error {
	return r.db.Where("nginx_id = ?", nginxID).Delete(&entities.NginxSecurity{}).Error
}
