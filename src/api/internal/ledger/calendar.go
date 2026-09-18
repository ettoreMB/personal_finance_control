package ledger

import (
	"time"
)

const saoPauloLocation = "America/Sao_Paulo"

func SaoPaulo() *time.Location {
	loc, err := time.LoadLocation(saoPauloLocation)
	if err != nil {
		return time.FixedZone("America/Sao_Paulo", -3*60*60)
	}
	return loc
}

func CivilDateIn(instant time.Time, loc *time.Location) time.Time {
	local := instant.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func TodayCivil() time.Time {
	return CivilDateIn(time.Now(), SaoPaulo())
}

func IsPastMonth(entryDate, today time.Time) bool {
	ey, em, _ := entryDate.UTC().Date()
	ty, tm, _ := today.UTC().Date()
	if ey != ty {
		return ey < ty
	}
	return em < tm
}

func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func AddMonthsClamped(anchor time.Time, months int) time.Time {
	y, m, d := anchor.UTC().Date()
	target := time.Date(y, m+time.Month(months), 1, 0, 0, 0, 0, time.UTC)
	last := lastDayOfMonth(target.Year(), target.Month())
	if d > last {
		d = last
	}
	return time.Date(target.Year(), target.Month(), d, 0, 0, 0, 0, time.UTC)
}

func SplitCents(total, n int) []int {
	parts := make([]int, n)
	base := total / n
	for i := 0; i < n; i++ {
		parts[i] = base
	}
	parts[n-1] += total % n
	return parts
}
