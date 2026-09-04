package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
	"github.com/nhassl3/IpBuild-backend/internal/repository/postgres"
	"github.com/nhassl3/IpBuild-backend/pkg/logger"
	"github.com/nhassl3/IpBuild-backend/pkg/mailer"
	"github.com/nhassl3/IpBuild-backend/pkg/minio"
)

const (
	// resumeEmailTTL is how long the résumé link embedded in the owner
	// notification email stays valid.
	resumeEmailTTL = 7 * 24 * time.Hour
	// resumeViewTTL is how long a résumé link returned to an admin via the API
	// stays valid.
	resumeViewTTL = 15 * time.Minute
)

type VacanciesService struct {
	repo        postgres.Vacancies
	mailer      mailer.Notifier
	minioClient minio.ByteStorage
	log         logger.Logger
}

func NewVacanciesService(repo postgres.Vacancies, mailer mailer.Notifier, minioClient minio.ByteStorage, log logger.Logger) *VacanciesService {
	return &VacanciesService{repo: repo, mailer: mailer, minioClient: minioClient, log: log}
}

func (s *VacanciesService) List(ctx context.Context, limit, offset int32) (*domain.VacanciesWithJd, error) {
	vacancies, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.List: %w", err)
	}
	return vacancies, nil
}

func (s *VacanciesService) GetVacancy(ctx context.Context, vacancyId string) (*domain.VacancyWithJd, error) {
	vacancy, err := s.repo.GetVacancy(ctx, vacancyId)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.GetVacancy: %w", err)
	}
	return vacancy, nil
}

func (s *VacanciesService) Create(ctx context.Context, params *domain.CreateVacancyInput) (*domain.Vacancy, error) {
	vacancy, err := s.repo.CreateVacancy(ctx, params)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return nil, domain.ErrVacancyAlreadyExists
			}
		}
		return nil, fmt.Errorf("vacancies_service.Create: %w", err)
	}
	return vacancy, nil
}

func (s *VacanciesService) Update(ctx context.Context, vacancyId string, updVacancy *domain.UpdatedVacancyInput) (*domain.Vacancy, error) {
	if err := s.repo.UpdateVacancy(ctx, vacancyId, updVacancy); err != nil {
		return nil, fmt.Errorf("vacancies_service.Update: %w", err)
	}
	vacancy, err := s.repo.GetVacancy(ctx, vacancyId)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.Update: fetch after update: %w", err)
	}
	return &vacancy.Vacancy, nil
}

func (s *VacanciesService) Delete(ctx context.Context, vacancyId string) error {
	if err := s.repo.DeleteVacancy(ctx, vacancyId); err != nil {
		return fmt.Errorf("vacancies_service.Delete: %w", err)
	}
	return nil
}

func (s *VacanciesService) ListJd(ctx context.Context, limit, offset int32) (*domain.JobDirections, error) {
	jd, err := s.repo.ListJd(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.ListJd: %w", err)
	}
	return jd, nil
}

func (s *VacanciesService) GetJd(ctx context.Context, jdId int64) (*domain.JobDirection, error) {
	jd, err := s.repo.GetJd(ctx, jdId)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.GetJd: %w", err)
	}
	return jd, nil
}

func (s *VacanciesService) CreateJd(ctx context.Context, params *domain.CreateJobDirectionInput) (*domain.JobDirection, error) {
	jd, err := s.repo.CreateJd(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.CreateJd: %w", err)
	}
	return jd, nil
}

func (s *VacanciesService) UpdateJd(ctx context.Context, jdId int64, updJd *domain.UpdateJobDirectionInput) (*domain.JobDirection, error) {
	jd, err := s.repo.UpdateJd(ctx, jdId, updJd)
	if err != nil {
		return nil, fmt.Errorf("vacancies_service.UpdateJd: %w", err)
	}
	return jd, nil
}

func (s *VacanciesService) DeleteJd(ctx context.Context, jdId int64) error {
	if err := s.repo.RemoveJd(ctx, jdId); err != nil {
		return fmt.Errorf("vacancies_service.DeleteJd: %w", err)
	}
	return nil
}

// Respond saves the applicant's form to the DB and asynchronously notifies the
// owner by email. SMTP errors are logged but do not fail the request.
func (s *VacanciesService) Respond(ctx context.Context, vacancyId string, applicantsForm *domain.ApplicantsFormInput, fileInput *domain.FileUploadInput) error {
	// Detect the real content type from the bytes (the client header is not
	// trusted) and validate size/type.
	contentType, err := minio.ResolveContentType(fileInput.FileData)
	if err != nil {
		return fmt.Errorf("vacancies_service.Respond.minio: failed to validate uploaded file: %w", err)
	}

	vacancy, err := s.repo.GetVacancy(ctx, vacancyId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrVacancyNotExists
		}
		return fmt.Errorf("vacancies_service.Respond: failed to get vacancy %w", err)
	}

	objectName := minio.GenerateObjectName("resumes/users", applicantsForm.Email, contentType)

	if _, err := s.minioClient.Upload(
		ctx, objectName, contentType, bytes.NewReader(fileInput.FileData), int64(len(fileInput.FileData)),
	); err != nil {
		return fmt.Errorf("vacancies_service.Respond.minio: failed to upload file: %w", err)
	}

	if _, err := s.repo.RespondToVacancy(ctx, vacancyId, objectName, applicantsForm); err != nil {
		// Don't leave an orphaned object behind if persisting the response failed.
		if delErr := s.minioClient.Delete(ctx, objectName); delErr != nil {
			s.log.Error("cleanup orphaned object failed",
				logger.Op("Respond"), logger.String("object", objectName), logger.Err(delErr))
		}
		return fmt.Errorf("vacancies_service.Respond: %w", err)
	}

	// A failed notification link must not fail the whole request.
	resumeURL, err := s.minioClient.PresignedURL(ctx, objectName, resumeEmailTTL)
	if err != nil {
		s.log.Warn("presign resume for email failed",
			logger.Op("Respond"), logger.String("object", objectName), logger.Err(err))
	}

	go func() {
		_ = s.mailer.NotifyNewApplicant(ctx, vacancy.Name, resumeURL, applicantsForm)
		_ = s.mailer.NotifyUserAboutVacancy(ctx, vacancy.Name, applicantsForm.Email)
	}()

	return nil
}

// presignResume replaces the stored object key on rv with a temporary signed
// URL. On failure the link is cleared rather than failing the whole listing.
func (s *VacanciesService) presignResume(ctx context.Context, rv *domain.RespondVacancy) {
	if rv.ResumeUrl == "" {
		return
	}
	url, err := s.minioClient.PresignedURL(ctx, rv.ResumeUrl, resumeViewTTL)
	if err != nil {
		s.log.Warn("presign resume failed",
			logger.Op("presignResume"), logger.String("object", rv.ResumeUrl), logger.Err(err))
		rv.ResumeUrl = ""
		return
	}
	rv.ResumeUrl = url
}

func (s *VacanciesService) GetRespondVacancies(ctx context.Context) (*domain.RespondVacancies, error) {
	respondVacancies, err := s.repo.GetRespondVacancies(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRespondVacanciesNotExists
		}
		return nil, fmt.Errorf("vacancies_service.GetRespondVacancies: %w", err)
	}
	for i := range respondVacancies.RespondVacancies {
		s.presignResume(ctx, &respondVacancies.RespondVacancies[i])
	}
	return respondVacancies, nil
}

func (s *VacanciesService) GetRespondVacancy(ctx context.Context, respondVacancyId string) (*domain.RespondVacancy, error) {
	respondVacancy, err := s.repo.GetRespondVacancy(ctx, respondVacancyId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRespondVacancyNotExists
		}
		return nil, fmt.Errorf("vacancies_service.GetRespondVacancy: %w", err)
	}
	s.presignResume(ctx, respondVacancy)
	return respondVacancy, nil
}
