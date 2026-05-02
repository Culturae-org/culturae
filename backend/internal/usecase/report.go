// backend/internal/usecase/report.go

package usecase

import (
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/repository"
	"github.com/Culturae-org/culturae/internal/service"
	"time"

	"github.com/google/uuid"
)

type ReportUsecase struct {
	reportRepo repository.ReportRepositoryInterface
	gameRepo   repository.GameRepositoryInterface
	wsService  service.WebSocketServiceInterface
}

func NewReportUsecase(
	reportRepo repository.ReportRepositoryInterface,
	gameRepo repository.GameRepositoryInterface,
	wsService service.WebSocketServiceInterface,
) *ReportUsecase {
	return &ReportUsecase{
		reportRepo: reportRepo,
		gameRepo:   gameRepo,
		wsService:  wsService,
	}
}

// -----------------------------------------------
// Report Usecase Methods
//
// - CreateReportFromGame
//
// -----------------------------------------------

func (u *ReportUsecase) CreateReportFromGame(userID uuid.UUID, gamePublicID string, questionNumber int, req model.CreateGameReportRequest) (*model.QuestionReport, error) {
	game, err := u.gameRepo.GetGameByPublicID(gamePublicID)
	if err != nil {
		return nil, err
	}

	gq, err := u.gameRepo.GetGameQuestionByOrder(game.ID, questionNumber)
	if err != nil {
		return nil, err
	}

	report := &model.QuestionReport{
		ID:             uuid.New(),
		UserID:         userID,
		GameQuestionID: &gq.ID,
		QuestionID:     gq.QuestionID,
		Reason:         req.Reason,
		Message:        req.Message,
		Status:         model.ReportStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := u.reportRepo.CreateReport(report); err != nil {
		return nil, err
	}

	u.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "report_created",
		Data: map[string]interface{}{
			keyReason: req.Reason,
		},
		EntityType: "report",
		EntityID:   report.ID.String(),
		ActionURL:  "/reports/" + report.ID.String(),
	})

	return report, nil
}
