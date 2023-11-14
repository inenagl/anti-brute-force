package ratelimit

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIncrement(t *testing.T) {
	// Устанавливаем ёмкость бакета и скорость изливания в ед/сек.
	capacity := 5
	rate := 10.0

	b := NewBucket(capacity, rate)

	// Успешно наполняеем бакет полностью.
	for i := 1; i <= capacity; i++ {
		require.True(t, b.Increment())
	}

	// Следующие попытки наполнения не приведут к успеху, т.к. мы выбрали всю ёмкость.
	for i := 1; i <= capacity; i++ {
		require.False(t, b.Increment())
	}

	// Ждём время, необходимое для удаления из бакета одной единицы.
	time.Sleep(time.Microsecond * time.Duration(math.Ceil(1/rate*1000000)))

	// Инкремент должен пройти один раз
	require.True(t, b.Increment())
	require.False(t, b.Increment())

	// Ждём время, необходимое для полного опустошения бакета.
	time.Sleep(time.Microsecond * time.Duration(math.Ceil(float64(capacity)/rate*1000000)))

	// Теперь снова заполняем бакет.
	for i := 1; i <= capacity; i++ {
		require.True(t, b.Increment())
	}

	// А здесь бакет снова переполнен.
	for i := 1; i <= capacity; i++ {
		require.False(t, b.Increment())
	}
}
