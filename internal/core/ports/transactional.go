package ports

import "gorm.io/gorm"

type Transactional[T any] interface {
	WithTx(tx *gorm.DB) T
}
