package dto

type Subscription struct {
	Id           string `json:"id" gorm:"primaryKey"`
	TranslatorIp string `json:"translatorIp,omitempty" gorm:"index:translator_subscriber_idx,unique"`
	SubscriberIp string `json:"subscriberIp,omitempty" gorm:"index:translator_subscriber_idx,unique"`
	Protocol     string `json:"protocol,omitempty"`
}
