package repositories

import (
	"translators/src/dto"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SubscribeRepository struct {
	Db  *gorm.DB
	Log *zap.Logger
}

func (repo SubscribeRepository) FindMany(ip string) ([]dto.Subscription, error) {
	subscriptions := []dto.Subscription{}

	err := repo.Db.Model(&dto.Subscription{}).Where("sub_ip = ?", ip).Find(&subscriptions).Error
	return subscriptions, err
}

func (repo SubscribeRepository) Create(subscription dto.Subscription) error {
	return repo.Db.Create(&subscription).Error
}

func (repo SubscribeRepository) DeleteOne(subIp string, translatorIp string) error {
	err := repo.Db.Where("sub_ip = ? AND translator_ip = ?", subIp, translatorIp).Delete(&dto.Subscription{}).Error
	return err
}

func (repo SubscribeRepository) DeleteMany(subIp string) error {
	err := repo.Db.Where("sub_ip = ?", subIp).Delete(&dto.Subscription{}).Error
	return err
}
