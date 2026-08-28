package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nhassl3/IpBuild-backend/internal/db"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

type VacanciesRepo struct {
	db *db.Store
}

func NewVacanciesRepo(db *db.Store) *VacanciesRepo {
	return &VacanciesRepo{
		db: db,
	}
}

func (r *VacanciesRepo) List(ctx context.Context, limit, offset int32) (*domain.VacanciesWithJd, error) {
	if limit == 0 {
		limit = 4
	}
	vacancies, err := r.db.GetVacancies(ctx, db.GetVacanciesParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("vacancies_repo.List: failed to load vacancies: %w", err)
	}
	return mapVacancies(vacancies), nil
}

func (r *VacanciesRepo) GetVacancy(ctx context.Context, vacancyId string) (*domain.VacancyWithJd, error) {
	vacancy, err := r.db.GetVacancy(ctx, db.GetVacancyParams{
		ID: uuidPtr2Nullable(vacancyId),
	})
	if err != nil {
		return nil, fmt.Errorf("vacancies_repo.GetVacancy: failed to load vacancy: %w", err)
	}
	return new(mapVacancyWithJd(vacancy)), nil
}

func (r *VacanciesRepo) CreateVacancy(ctx context.Context, params *domain.CreateVacancyInput) (*domain.Vacancy, error) {
	payDay := 0.0
	if params.PayDay != nil {
		payDay = *params.PayDay
	}
	vacancy, err := r.db.CreateVacancy(ctx, db.CreateVacancyParams{
		Jd:          params.Jd,
		Name:        stringToNullable(params.Name),
		Description: stringToNullable(params.Description),
		RequiredExp: stringPtrToNullable(params.RequiredExp),
		PayDay:      payDay,
		Skills:      params.Skills,
	})
	if err != nil {
		return nil, fmt.Errorf("vacancies_repo.Create: failed to create vacancy: %w", err)
	}
	return new(mapVacancy(vacancy)), nil
}

func (r *VacanciesRepo) UpdateVacancy(ctx context.Context, vacancyId string, updVacancy *domain.UpdatedVacancyInput) error {
	if updVacancy == nil {
		return fmt.Errorf("vacancies_repo.Update: vacancies input is nil")
	}
	return r.db.ExecTx(ctx, func(q *db.Queries) error {
		var (
			updateVacancyParams db.UpdateVacancyParams
			fnErr               error
		)

		vacancy, fnErr := q.GetVacancy(ctx, db.GetVacancyParams{
			ID: uuidPtr2Nullable(vacancyId),
		})
		if fnErr != nil {
			return fmt.Errorf("vacancies_repo.Update: failed to load vacancy id: %w", fnErr)
		}

		updateVacancyParams = db.UpdateVacancyParams{
			ID:          uuidPtr2Nullable(vacancyId),
			Jd:          vacancy.Jd,
			NewName:     vacancy.Name.String,
			Description: vacancy.Description,
			RequiredExp: vacancy.RequiredExp,
			PayDay:      vacancy.PayDay,
			Skills:      vacancy.Skills,
		}
		if updVacancy.Jd != nil && *updVacancy.Jd != 0 {
			updateVacancyParams.Jd = *updVacancy.Jd
		}
		if updVacancy.Skills != nil {
			updateVacancyParams.Skills = updVacancy.Skills
		}
		if updVacancy.Name != nil && *updVacancy.Name != "" {
			updateVacancyParams.NewName = *updVacancy.Name
		}
		if updVacancy.Description != nil && *updVacancy.Description != "" {
			updateVacancyParams.Description = stringPtrToNullable(updVacancy.Description)
		}
		if updVacancy.RequiredExp != nil && *updVacancy.RequiredExp != "" {
			updateVacancyParams.RequiredExp = stringPtrToNullable(updVacancy.RequiredExp)
		}
		if updVacancy.PayDay != nil && *updVacancy.PayDay > 0 {
			updateVacancyParams.PayDay = *updVacancy.PayDay
		}

		if _, fnErr = q.UpdateVacancy(ctx, updateVacancyParams); fnErr != nil {
			return fmt.Errorf("vacancies_repo.Update: failed to update vacancy: %w", fnErr)
		}

		return nil
	})
}

func (r *VacanciesRepo) DeleteVacancy(ctx context.Context, vacancyId string) error {
	return r.db.RemoveVacancy(ctx, db.RemoveVacancyParams{
		ID: uuidPtr2Nullable(vacancyId),
	})
}

func (r *VacanciesRepo) ListJd(ctx context.Context, limit, offset int32) (*domain.JobDirections, error) {
	if limit == 0 {
		limit = 4
	}
	jds, err := r.db.GetJDs(ctx, db.GetJDsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("vacancies_repo.ListJd: failed to load job directions: %w", err)
	}
	return mapJobDirections(jds), nil
}

func (r *VacanciesRepo) GetJd(ctx context.Context, jdId int32) (*domain.JobDirection, error) {
	jd, err := r.db.GetJD(ctx, int64(jdId))
	if err != nil {
		return nil, fmt.Errorf("vacancies_repo.GetJd: failed to load job direction: %w", err)
	}
	return new(mapJobDirection(jd)), nil
}

func (r *VacanciesRepo) CreateJd(ctx context.Context, params *domain.CreateJobDirectionInput) (*domain.JobDirection, error) {
	tags := params.Tags
	if tags == nil {
		tags = []string{}
	}
	jd, err := r.db.CreateJobDirection(ctx, db.CreateJobDirectionParams{
		Name:        params.Name,
		Tags:        tags,
		Description: params.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("vacancies_repo.CreateJd: failed to create job direction: %w", err)
	}
	return new(mapJobDirection(jd)), nil
}

func (r *VacanciesRepo) UpdateJd(ctx context.Context, jdId int32, params *domain.UpdateJobDirectionInput) (*domain.JobDirection, error) {
	jd, err := r.db.UpdateJobDirection(ctx, db.UpdateJobDirectionParams{
		ID:          int64(jdId),
		Name:        stringPtrToNullable(params.Name),
		Tags:        params.Tags,
		Description: stringPtrToNullable(params.Description),
	})
	if err != nil {
		return nil, fmt.Errorf("vacancies_repo.UpdateJd: failed to update job direction: %w", err)
	}
	return new(mapJobDirection(jd)), nil
}

func (r *VacanciesRepo) RemoveJd(ctx context.Context, jdId int32) error {
	if err := r.db.RemoveJobDirection(ctx, int64(jdId)); err != nil {
		// 23503 — foreign key violation: vacancies still reference this direction.
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23503" {
				return domain.ErrDirectionHasVacancies
			}
		}
		return fmt.Errorf("vacancies_repo.RemoveJd: failed to remove job direction: %w", err)
	}
	return nil
}

func (r *VacanciesRepo) RespondToVacancy(ctx context.Context, vacancyId, objectName string, applicantsForm *domain.ApplicantsFormInput) (string, error) {
	userRespondId, err := r.db.RespondToVacancy(ctx, db.RespondToVacancyParams{
		FullName:    applicantsForm.FullName,
		Email:       applicantsForm.Email,
		PhoneNumber: stringToNullable(applicantsForm.PhoneNumber),
		City:        stringToNullable(applicantsForm.City),
		Exp:         stringToNullable(applicantsForm.Exp),
		Description: stringToNullable(applicantsForm.Description),
		Resume:      stringToNullable(objectName),
		VacancyID:   string2UUID(vacancyId),
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return "", domain.ErrRespondAlreadyExists
			}
		}
		return "", fmt.Errorf("vacancies_repo.RespondToVacancy: %w", err)
	}
	return uuid2String(userRespondId), nil
}

func (r *VacanciesRepo) GetRespondVacancies(ctx context.Context) (*domain.RespondVacancies, error) {
	respondVacancies, err := r.db.GetRespondVacancies(ctx, db.GetRespondVacanciesParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("vacancies_repo.GetRespondVacancies: %w", err)
	}
	return mapRespondVacancies(respondVacancies), nil
}

func (r *VacanciesRepo) GetRespondVacancy(ctx context.Context, respondVacancyId string) (*domain.RespondVacancy, error) {
	respondVacancy, err := r.db.GetRespondVacancy(ctx, string2UUID(respondVacancyId))
	if err != nil {
		return nil, fmt.Errorf("vacancies_repo.GetRespondVacancy: %w", err)
	}
	return new(mapRespondVacancy(respondVacancy)), nil
}

func mapVacancies(vacancies []db.VacancyWithJd) *domain.VacanciesWithJd {
	domainVacancies := make([]domain.VacancyWithJd, len(vacancies))
	for i := range vacancies {
		domainVacancies[i] = mapVacancyWithJd(vacancies[i])
	}
	return &domain.VacanciesWithJd{
		VacanciesWithJd: domainVacancies,
	}
}

func mapVacancyWithJd(v db.VacancyWithJd) domain.VacancyWithJd {
	return domain.VacancyWithJd{
		Vacancy: domain.Vacancy{
			UUID:        uuid2String(v.ID),
			Name:        v.Name.String,
			Description: v.Description.String,
			RequiredExp: v.RequiredExp.String,
			PayDay:      v.PayDay,
			Skills:      v.Skills,
			CreatedAt:   pgTimeTZ(v.CreatedAt, time.UTC),
			UpdatedAt:   pgTimeTZ(v.UpdatedAt, time.UTC),
		},
		JobDirection: domain.JobDirection{
			Id:            v.Jd,
			JdName:        v.JdName,
			JdTags:        v.JdTags,
			JdDescription: v.JdDescription,
		},
	}
}

func mapVacancy(v db.Vacancy) domain.Vacancy {
	return domain.Vacancy{
		UUID:        uuid2String(v.ID),
		Name:        v.Name.String,
		Description: v.Description.String,
		RequiredExp: v.RequiredExp.String,
		PayDay:      v.PayDay,
		Skills:      v.Skills,
		CreatedAt:   pgTimeTZ(v.CreatedAt, time.UTC),
		UpdatedAt:   pgTimeTZ(v.UpdatedAt, time.UTC),
	}
}

func mapJobDirections(jds []db.JobDirection) *domain.JobDirections {
	domainJds := make([]domain.JobDirection, len(jds))
	for i := range jds {
		domainJds[i] = mapJobDirection(jds[i])
	}
	return &domain.JobDirections{
		JobDirections: domainJds,
	}
}

func mapJobDirection(jd db.JobDirection) domain.JobDirection {
	return domain.JobDirection{
		Id:            jd.ID,
		JdName:        jd.Name,
		JdTags:        jd.Tags,
		JdDescription: jd.Description,
	}
}

func mapRespondVacancies(respondVacancies []db.UserRespond) *domain.RespondVacancies {
	domainRespondVacancies := make([]domain.RespondVacancy, len(respondVacancies))
	for i := range respondVacancies {
		domainRespondVacancies[i] = mapRespondVacancy(respondVacancies[i])
	}
	return &domain.RespondVacancies{
		RespondVacancies: domainRespondVacancies,
	}
}

func mapRespondVacancy(v db.UserRespond) domain.RespondVacancy {
	return domain.RespondVacancy{
		UUID:        uuid2String(v.ID),
		FullName:    v.FullName,
		Email:       v.Email,
		PhoneNumber: v.PhoneNumber.String,
		City:        v.City.String,
		Exp:         v.Exp.String,
		Description: v.Description.String,
		// Raw object key; the service turns it into a presigned URL on read.
		ResumeUrl: v.Resume.String,
		VacancyId: uuid2String(v.VacancyID),
		CreatedAt: pgTimeTZ(v.CreatedAt, time.UTC),
	}
}
