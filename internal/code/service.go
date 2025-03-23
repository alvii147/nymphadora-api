package code

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/internal/config"
	"github.com/alvii147/nymphadora-api/internal/templatesmanager"
	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/mailclient"
	"github.com/alvii147/nymphadora-api/pkg/piston"
	"github.com/alvii147/nymphadora-api/pkg/random"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FrontendCodeSpaceInvitationRoute is the frontend route for code space invitation.
const FrontendCodeSpaceInvitationRoute = "/code/space/%s/invitation/%s"

// Service performs all code-space-related business logic.
//
//go:generate mockgen -package=codemocks -source=$GOFILE -destination=./mocks/service.go
type Service interface {
	CreateCodeSpace(
		ctx context.Context,
		language string,
	) (*CodeSpace, *CodeSpaceAccess, error)
	ListCodeSpaces(
		ctx context.Context,
	) ([]*CodeSpace, []*CodeSpaceAccess, error)
	GetCodespace(
		ctx context.Context,
		name string,
	) (*CodeSpace, *CodeSpaceAccess, error)
	UpdateCodeSpace(
		ctx context.Context,
		name string,
		contents *string,
	) (*CodeSpace, *CodeSpaceAccess, error)
	DeleteCodeSpace(
		ctx context.Context,
		name string,
	) error
	RunCodeSpace(
		ctx context.Context,
		name string,
	) (*api.PistonExecuteResponse, error)
	ListCodespaceUsers(
		ctx context.Context,
		name string,
	) ([]*auth.User, []*CodeSpaceAccess, error)
	SendCodeSpaceInvitationMail(
		ctx context.Context,
		email string,
		data templatesmanager.CodeSpaceInvitationEmailTemplateData,
	) error
	InviteCodeSpaceUser(
		ctx context.Context,
		name string,
		inviteeEmail string,
		accessLevel CodeSpaceAccessLevel,
	) error
	RemoveCodeSpaceUser(
		ctx context.Context,
		name string,
		inviteeUUID string,
	) error
	AcceptCodeSpaceUserInvitation(
		ctx context.Context,
		name string,
		token string,
	) (*CodeSpace, *CodeSpaceAccess, error)
}

// service implements Service.
type service struct {
	config         *config.Config
	timeProvider   timekeeper.Provider
	dbPool         *pgxpool.Pool
	crypto         cryptocore.Crypto
	mailClient     mailclient.Client
	tmplManager    templatesmanager.Manager
	pistonClient   piston.Client
	repository     Repository
	authRepository auth.Repository
}

// NewService returns a new service.
func NewService(
	config *config.Config,
	timeProvider timekeeper.Provider,
	dbPool *pgxpool.Pool,
	crypto cryptocore.Crypto,
	mailClient mailclient.Client,
	tmplManager templatesmanager.Manager,
	pistonClient piston.Client,
	repo Repository,
	authRepository auth.Repository,
) *service {
	return &service{
		config:         config,
		timeProvider:   timeProvider,
		dbPool:         dbPool,
		crypto:         crypto,
		mailClient:     mailClient,
		tmplManager:    tmplManager,
		pistonClient:   pistonClient,
		repository:     repo,
		authRepository: authRepository,
	}
}

// GenerateCodeSpaceName generates a random name for a code space.
func (svc *service) GenerateCodeSpaceName() (string, error) {
	adjectiveIdx, err := random.GenerateInt64(int64(len(Adjectives)))
	if err != nil {
		return "", errutils.FormatError(err)
	}

	pokemonIdx, err := random.GenerateInt64(int64(len(Pokemon)))
	if err != nil {
		return "", errutils.FormatError(err)
	}

	programmingTermIdx, err := random.GenerateInt64(int64(len(ProgrammingTerms)))
	if err != nil {
		return "", errutils.FormatError(err)
	}

	harryPotterCharacterIdx, err := random.GenerateInt64(int64(len(HarryPotterCharacters)))
	if err != nil {
		return "", errutils.FormatError(err)
	}

	fileExtensionIdx, err := random.GenerateInt64(int64(len(FileExtensions)))
	if err != nil {
		return "", errutils.FormatError(err)
	}

	name := strings.Join(
		[]string{
			Adjectives[adjectiveIdx],
			Pokemon[pokemonIdx],
			ProgrammingTerms[programmingTermIdx],
			HarryPotterCharacters[harryPotterCharacterIdx],
			FileExtensions[fileExtensionIdx],
		},
		"-",
	)

	return name, nil
}

// CreateCodeSpace creates a new code space.
func (svc *service) CreateCodeSpace(
	ctx context.Context,
	language string,
) (*CodeSpace, *CodeSpaceAccess, error) {
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	languageConfig, ok := CodingLanguageConfig[language]
	if !ok {
		return nil, nil, errutils.FormatErrorf(errutils.ErrCodeSpaceUnsupportedLanguage, "unknown language %s", language)
	}

	templateFilePath := fmt.Sprintf("_codetemplates/%s/%s", language, languageConfig.fileName)
	templateFileBytes, err := CodeTemplatesFS.ReadFile(templateFilePath)
	if err != nil {
		return nil, nil, errutils.FormatErrorf(nil, "codeTemplatesFS.ReadFile failed to read %s", templateFilePath)
	}

	name, err := svc.GenerateCodeSpaceName()
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpace := &CodeSpace{
		Name:       name,
		AuthorUUID: &userUUID,
		Language:   language,
		Contents:   string(templateFileBytes),
	}

	codeSpace, err = svc.repository.CreateCodeSpace(ctx, dbConn, codeSpace)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	codeSpaceAccess := &CodeSpaceAccess{
		UserUUID:    userUUID,
		CodeSpaceID: codeSpace.ID,
		Level:       CodeSpaceAccessLevelReadWrite,
	}

	codeSpaceAccess, err = svc.repository.CreateOrUpdateCodeSpaceAccess(ctx, dbConn, codeSpaceAccess)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	return codeSpace, codeSpaceAccess, nil
}

// ListCodeSpaces lists all code spaces accessible to the currently authenticated user.
func (svc *service) ListCodeSpaces(
	ctx context.Context,
) ([]*CodeSpace, []*CodeSpaceAccess, error) {
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpaces, codeSpaceAccesses, err := svc.repository.ListCodeSpaces(ctx, dbConn, userUUID)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	return codeSpaces, codeSpaceAccesses, nil
}

// GetCodespace gets a given code space for the currently authenticated user.
func (svc *service) GetCodespace(
	ctx context.Context,
	name string,
) (*CodeSpace, *CodeSpaceAccess, error) {
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpace, codeSpaceAccess, err := svc.repository.GetCodeSpaceWithAccessByName(ctx, dbConn, userUUID, name)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = errutils.FormatError(errutils.ErrCodeSpaceNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return nil, nil, err
	}

	return codeSpace, codeSpaceAccess, nil
}

// UpdateCodeSpace updates a given code space.
func (svc *service) UpdateCodeSpace(
	ctx context.Context,
	name string,
	contents *string,
) (*CodeSpace, *CodeSpaceAccess, error) {
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpace, codeSpaceAccess, err := svc.repository.GetCodeSpaceWithAccessByName(
		ctx,
		dbConn,
		userUUID,
		name,
	)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = errutils.FormatError(errutils.ErrCodeSpaceNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return nil, nil, err
	}

	if codeSpaceAccess.Level < CodeSpaceAccessLevelReadWrite {
		return nil, nil, errutils.FormatError(errutils.ErrCodeSpaceAccessDenied)
	}

	codeSpace, err = svc.repository.UpdateCodeSpace(
		ctx,
		dbConn,
		codeSpace.ID,
		contents,
	)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	return codeSpace, codeSpaceAccess, nil
}

// DeleteCodeSpace deletes a code space.
func (svc *service) DeleteCodeSpace(
	ctx context.Context,
	name string,
) error {
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		return errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpace, codeSpaceAccess, err := svc.repository.GetCodeSpaceWithAccessByName(
		ctx,
		dbConn,
		userUUID,
		name,
	)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = errutils.FormatError(errutils.ErrCodeSpaceNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return err
	}

	if codeSpaceAccess.Level < CodeSpaceAccessLevelReadWrite {
		return errutils.FormatError(errutils.ErrCodeSpaceAccessDenied)
	}

	err = svc.repository.DeleteCodeSpace(ctx, dbConn, codeSpace.ID)
	if err != nil {
		return errutils.FormatError(err)
	}

	return nil
}

// RunCodeSpace runs the code in a code space.
func (svc *service) RunCodeSpace(
	ctx context.Context,
	name string,
) (*api.PistonExecuteResponse, error) {
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpace, _, err := svc.repository.GetCodeSpaceWithAccessByName(ctx, dbConn, userUUID, name)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = errutils.FormatError(errutils.ErrCodeSpaceNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return nil, err
	}

	languageConfig, ok := CodingLanguageConfig[codeSpace.Language]
	if !ok {
		return nil, errutils.FormatErrorf(nil, "unknown language %s", codeSpace.Language)
	}

	encoding := api.PistonFileEncoding
	req := &api.PistonExecuteRequest{
		Language: codeSpace.Language,
		Version:  languageConfig.version,
		Files: []api.PistonFile{
			{
				Name:     &languageConfig.fileName,
				Content:  codeSpace.Contents,
				Encoding: &encoding,
			},
		},
	}

	resp, err := svc.pistonClient.Execute(req)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	return resp, nil
}

// ListCodespaceUsers lists users with access to a code space.
func (svc *service) ListCodespaceUsers(
	ctx context.Context,
	name string,
) ([]*auth.User, []*CodeSpaceAccess, error) {
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpace, codeSpaceAccess, err := svc.repository.GetCodeSpaceWithAccessByName(
		ctx,
		dbConn,
		userUUID,
		name,
	)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = errutils.FormatError(errutils.ErrCodeSpaceNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return nil, nil, err
	}

	if codeSpaceAccess.Level < CodeSpaceAccessLevelReadWrite {
		return nil, nil, errutils.FormatError(errutils.ErrCodeSpaceAccessDenied)
	}

	users, codeSpaceAccesses, err := svc.repository.ListUsersWithCodeSpaceAccess(ctx, dbConn, codeSpace.ID)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	return users, codeSpaceAccesses, nil
}

// SendCodeSpaceInvitationMail sends the code space invitation email.
func (svc *service) SendCodeSpaceInvitationMail(
	ctx context.Context,
	email string,
	data templatesmanager.CodeSpaceInvitationEmailTemplateData,
) error {
	textTmpl, htmlTmpl, err := svc.tmplManager.Load(templatesmanager.CodeSpaceInvitationEmailTemplateName)
	if err != nil {
		return errutils.FormatError(err)
	}

	err = svc.mailClient.Send([]string{email}, templatesmanager.CodeSpaceInvitationEmailSubject, textTmpl, htmlTmpl, data)
	if err != nil {
		return errutils.FormatError(err)
	}

	return nil
}

// InviteCodeSpaceUser invites a user to a code space by email.
func (svc *service) InviteCodeSpaceUser(
	ctx context.Context,
	name string,
	inviteeEmail string,
	accessLevel CodeSpaceAccessLevel,
) error {
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		return errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpace, codeSpaceAccess, err := svc.repository.GetCodeSpaceWithAccessByName(
		ctx,
		dbConn,
		userUUID,
		name,
	)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = errutils.FormatError(errutils.ErrCodeSpaceNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return err
	}

	if codeSpaceAccess.Level < CodeSpaceAccessLevelReadWrite {
		return errutils.FormatError(errutils.ErrCodeSpaceAccessDenied)
	}

	token, err := svc.crypto.CreateCodeSpaceInvitationJWT(
		userUUID,
		inviteeEmail,
		codeSpace.ID,
		int(accessLevel),
	)
	if err != nil {
		return errutils.FormatError(err)
	}

	invitationURL := fmt.Sprintf(
		svc.config.FrontendBaseURL+FrontendCodeSpaceInvitationRoute,
		codeSpace.Name,
		token,
	)
	data := templatesmanager.CodeSpaceInvitationEmailTemplateData{
		InvitationURL: invitationURL,
	}

	err = svc.SendCodeSpaceInvitationMail(
		ctx,
		inviteeEmail,
		data,
	)
	if err != nil {
		return errutils.FormatError(err)
	}

	return nil
}

// RemoveCodeSpaceUser revokes a user's access to a code space.
func (svc *service) RemoveCodeSpaceUser(
	ctx context.Context,
	name string,
	inviteeUUID string,
) error {
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		return errutils.FormatError(err)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpace, codeSpaceAccess, err := svc.repository.GetCodeSpaceWithAccessByName(
		ctx,
		dbConn,
		userUUID,
		name,
	)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = errutils.FormatError(errutils.ErrCodeSpaceNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return err
	}

	if codeSpaceAccess.Level < CodeSpaceAccessLevelReadWrite {
		return errutils.FormatError(errutils.ErrCodeSpaceAccessDenied)
	}

	err = svc.repository.DeleteCodeSpaceAccess(ctx, dbConn, inviteeUUID, codeSpace.ID)
	if err != nil {
		return errutils.FormatError(err)
	}

	return nil
}

// AcceptCodeSpaceUserInvitation accepts a code space invitation from a given code space invitation JWT.
func (svc *service) AcceptCodeSpaceUserInvitation(
	ctx context.Context,
	name string,
	token string,
) (*CodeSpace, *CodeSpaceAccess, error) {
	claims, ok := svc.crypto.ValidateCodeSpaceInvitationJWT(token)
	if !ok {
		return nil, nil, errutils.FormatErrorf(errutils.ErrInvalidToken, "token %s", token)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, nil, errutils.FormatError(err, "svc.dbPool.Acquire failed")
	}
	defer dbConn.Release()

	codeSpace, err := svc.repository.GetCodeSpace(ctx, dbConn, claims.CodeSpaceID)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = errutils.FormatError(errutils.ErrCodeSpaceNotFound)
		default:
			err = errutils.FormatError(err)
		}

		return nil, nil, err
	}

	if name != codeSpace.Name {
		return nil, nil, errutils.FormatError(errutils.ErrCodeSpaceNotFound)
	}

	user, err := svc.authRepository.GetUserByEmail(ctx, dbConn, claims.InviteeEmail)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	codeSpaceAccess := &CodeSpaceAccess{
		UserUUID:    user.UUID,
		CodeSpaceID: codeSpace.ID,
		Level:       CodeSpaceAccessLevel(claims.AccessLevel),
	}

	codeSpaceAccess, err = svc.repository.CreateOrUpdateCodeSpaceAccess(ctx, dbConn, codeSpaceAccess)
	if err != nil {
		return nil, nil, errutils.FormatError(err)
	}

	return codeSpace, codeSpaceAccess, nil
}
