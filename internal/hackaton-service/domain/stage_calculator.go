package domain

import (
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
)

// ComputeStage вычисляет текущую стадию хакатона на основе временных меток.
// Логика соответствует спецификации из docs/rules/hackathon.md:36-45
func ComputeStage(now time.Time, hackathon *entity.Hackathon) string {
	// 1. DRAFT: published_at == null
	if hackathon.PublishedAt == nil {
		return string(StageDraft)
	}

	// 2. FINISHED: result_published_at != null (приоритет!)
	if hackathon.ResultPublishedAt != nil {
		return string(StageFinished)
	}

	// 3. UPCOMING: now < registration_opens_at
	if hackathon.RegistrationOpensAt != nil && now.Before(*hackathon.RegistrationOpensAt) {
		return string(StageUpcoming)
	}

	// 4. REGISTRATION: registration_opens_at <= now < registration_closes_at
	if hackathon.RegistrationClosesAt != nil && now.Before(*hackathon.RegistrationClosesAt) {
		return string(StageRegistration)
	}

	// 5. PRESTART: registration_closes_at <= now < starts_at
	if hackathon.StartsAt != nil && now.Before(*hackathon.StartsAt) {
		return string(StagePreStart)
	}

	// 6. RUNNING: starts_at <= now < ends_at
	if hackathon.EndsAt != nil && now.Before(*hackathon.EndsAt) {
		return string(StageRunning)
	}

	// 7. JUDGING: ends_at <= now < judging_ends_at
	if hackathon.JudgingEndsAt != nil && now.Before(*hackathon.JudgingEndsAt) {
		return string(StageJudging)
	}

	// 8. FINISHED: now >= judging_ends_at (или все временные метки пройдены)
	return string(StageFinished)
}
