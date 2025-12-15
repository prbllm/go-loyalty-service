package model

// Order — заказ, полученный от доверенного источника
type Order struct {
	Number  string      // номер заказа
	Goods   []Good      // список купленных товаров
	Status  OrderStatus // статус расчета начисления
	Accrual *int64      // рассчитанные баллы к начислению, nil = нет начисления
}

type Good struct {
	Description string // наименование товара
	Price       int64  // цена оплаченного товара(в копейках)
}

type OrderStatus string

const (
	Registered OrderStatus = "REGISTERED" // заказ зарегистрирован, но не начисление не рассчитано;
	Processing OrderStatus = "PROCESSING" // расчёт начисления в процессе
	Processed  OrderStatus = "PROCESSED"  // расчёт начисления окончен
	Invalid    OrderStatus = "INVALID"    // заказ не принят к расчёту, и вознаграждение не будет начислено
)

// RewardRule — правило начисления за товар
type RewardRule struct {
	Match      string     `json:"match"`       // ключ поиска
	Reward     int64      `json:"reward"`      // размер вознаграждения
	RewardType RewardType `json:"reward_type"` // тип вознаграждения
}

type RewardType string

const (
	RewardTypePercent RewardType = "%"  // процент от стоимости товара
	RewardTypePoints  RewardType = "pt" // точное количество баллов
)
