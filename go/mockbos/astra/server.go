package astra

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"gitlab.com/fynbos/backend/providers/astra/external"
	"gitlab.com/fynbos/log"
	"gitlab.com/fynbos/mockbos/db"
	"gitlab.com/fynbos/mockbos/utils"
	"go.uber.org/zap"
)

var bearerTokenRegexp = regexp.MustCompile("Bearer .*")

type Server struct {
	q    *db.Queries
	conn *(pgx.Conn)
}

func New(conn *pgx.Conn) *Server {
	return &Server{conn: conn, q: db.New(conn)}
}

func (s *Server) Register(r chi.Router) {
	r.Post("/v1/user_intent", s.CreateUserIntent)
	r.Get("/v1/user_intent/{id}", s.GetUserIntent)
	r.Post("/v1/partner/identity/verification", s.CreateAccessTokenCode)
	r.Post("/v1/partner/identity/token", s.ExchangeCodeForAccessToken)
	r.Post("/v1/oath/token", s.RefreshToken)
	r.Post("/v1/cards", s.CreateUserCard)
	r.Get("/v1/cards/{id}", s.GetUserCard)
	r.Post("/v1/accounts/create", s.CreateUserAccount)
	r.Get("/v1/accounts/{id}", s.GetUserAccount)
	r.Post("/v1/routines/account-to-card", s.AccountToCard)
	r.Post("/v1/routines/card-to-account", s.CardToAccount)
	r.Get("/v1/transfers/{id}", s.GetTransfer)
}

func (s *Server) RegisterAdmin(r chi.Router) {
	r.Post("/convert_user_intent", s.AdminConvertIntentToUser)
}

func (s *Server) CreateUserIntent(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var payload external.CreateIntentReq
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("failed to unmarhsal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userIntent, err := s.q.CreateAstraIntent(r.Context(), db.CreateAstraIntentParams{
		Email:       payload.Email,
		Phone:       payload.Phone,
		FirstName:   payload.FirstName,
		LastName:    payload.LastName,
		Address1:    pgtype.Text{String: payload.Address1, Valid: true},
		Address2:    pgtype.Text{String: payload.Address2, Valid: true},
		City:        pgtype.Text{String: payload.City, Valid: true},
		State:       pgtype.Text{String: payload.State, Valid: true},
		PostalCode:  pgtype.Text{String: payload.PostalCode, Valid: true},
		DateOfBirth: pgtype.Text{String: payload.DateOfBirth, Valid: true},
		Ssn:         pgtype.Text{String: payload.SocialSecurity, Valid: true},
		IpAddress:   pgtype.Text{String: payload.IPAddress, Valid: true},
	})
	if err != nil {
		log.Error("failed to create intent", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.CreateIntentResp{
		ID: utils.BytesToUUID(userIntent.ID.Bytes),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetUserIntent(w http.ResponseWriter, r *http.Request) {
	parsedUUID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("Failed to parse user intent id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	intent, err := s.q.GetAstraIntent(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedUUID), Valid: true})
	if err != nil {
		log.Error("Failed to get user intent id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.Intent{
		ID:                 parsedUUID.String(),
		UserID:             utils.BytesToUUID(intent.UserID.Bytes),
		Email:              intent.Email,
		Phone:              intent.Phone,
		FirstName:          intent.FirstName,
		LastName:           intent.LastName,
		PreferredFirstName: intent.FirstName,
		PreferredLastName:  intent.LastName,
		Status:             "PENDING",
		KycType:            "UNVERIFIED",
	}
	if intent.UserID.Valid {
		usr, err := s.q.GetAstraUser(r.Context(), intent.UserID)
		if err != nil {
			log.Error("Failed to check for astra user", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		resp.Status = usr.Status.String
		resp.KycType = usr.KycType.String
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("Failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) AdminConvertIntentToUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var payload ConvertIntentToUser
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("failed to unmarhsal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		log.Error("failed to start transaction", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	parsedUUID, err := uuid.Parse(payload.ID)
	if err != nil {
		log.Error("failed to parse intent id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	q := s.q.WithTx(tx)
	intent, err := q.GetAstraIntent(ctx, pgtype.UUID{Bytes: [16]byte(parsedUUID), Valid: true})
	if err != nil {
		log.Error("failed to get intent", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	usr, err := q.CreateAstraUser(ctx, db.CreateAstraUserParams{
		Email:       intent.Email,
		Phone:       intent.Phone,
		FirstName:   intent.FirstName,
		LastName:    intent.LastName,
		Address1:    intent.Address1,
		Address2:    intent.Address2,
		City:        intent.City,
		State:       intent.State,
		PostalCode:  intent.PostalCode,
		DateOfBirth: intent.DateOfBirth,
		Ssn:         intent.Ssn,
		IpAddress:   intent.IpAddress,
		Status:      pgtype.Text{String: payload.Status, Valid: true},
		KycType:     pgtype.Text{String: payload.KycType, Valid: true},
	})
	if err != nil {
		log.Error("failed to create user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = q.UpdateAstraUserIntentUserID(ctx, db.UpdateAstraUserIntentUserIDParams{
		ID:     pgtype.UUID{Bytes: [16]byte(parsedUUID), Valid: true},
		UserID: usr.ID,
	})
	if err != nil {
		log.Error("failed to update intent with userID", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error("failed to commit transaction", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"userId": utils.BytesToUUID(usr.ID.Bytes),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateAccessTokenCode(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var payload external.GetVerificationTokenReq
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("failed to unmarhsal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	parsedIntentID, err := uuid.Parse(payload.UserIntentID)
	if err != nil {
		log.Error("failed to parse user intent id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	intent, err := s.q.GetAstraIntent(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedIntentID), Valid: true})
	if err != nil {
		log.Error("failed to get user intent id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := s.q.CreateAstraAccessToken(r.Context(), db.CreateAstraAccessTokenParams{
		UserID:    intent.UserID,
		ExpiresIn: pgtype.Int4{Int32: 36_000_000, Valid: true},
		TokenType: pgtype.Text{String: "ACCESS", Valid: true},
	})
	if err != nil {
		log.Error("failed to get user intent id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.VerificationTokenResp{
		Token: utils.BytesToUUID(token.ID.Bytes),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) ExchangeCodeForAccessToken(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Error("failed to parse x-www-form-urlencoded data", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	parsedTokenID, err := uuid.Parse(r.Form.Get("token"))
	if err != nil {
		log.Error("failed to parse token id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := s.q.GetAstraAccessToken(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedTokenID), Valid: true})
	if err != nil {
		log.Error("failed to get access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.AccessToken{
		AccessToken:  utils.BytesToUUID(token.ID.Bytes),
		RefreshToken: utils.BytesToUUID(token.RefreshToken.Bytes),
		TokenType:    token.TokenType.String,
		ExpiresIn:    int(token.ExpiresIn.Int32),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) RefreshToken(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Error("failed to parse x-www-form-urlencoded data", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	parsedRefreshTokenID, err := uuid.Parse(r.Form.Get("refresh_token"))
	if err != nil {
		log.Error("failed to parse refresh token id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := s.q.GetAstraAccessTokenByRefreshToken(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedRefreshTokenID), Valid: true})
	if err != nil {
		log.Error("failed to get access token by refresh token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.AccessToken{
		AccessToken:  utils.BytesToUUID(token.ID.Bytes),
		RefreshToken: utils.BytesToUUID(token.RefreshToken.Bytes),
		TokenType:    token.TokenType.String,
		ExpiresIn:    int(token.ExpiresIn.Int32),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateUserCard(w http.ResponseWriter, r *http.Request) {
	accessToken := getBearerToken(r)
	parsedAccessToken, err := uuid.Parse(accessToken)
	if err != nil {
		log.Error("failed to parse access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token, err := s.q.GetAstraAccessToken(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedAccessToken), Valid: true})
	if err != nil {
		log.Error("failed to get access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var payload external.CreateCardArgs
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("failed to unmarhsal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	lastFour := payload.CardNumber
	if len(payload.CardNumber) > 4 {
		lastFour = payload.CardNumber[len(payload.CardNumber)-4:]
	}

	firstSix := payload.CardNumber
	if len(payload.CardNumber) > 6 {
		firstSix = payload.CardNumber[0:6]
	}

	card, err := s.q.CreateAstraUserCard(r.Context(), db.CreateAstraUserCardParams{
		UserID:          token.UserID,
		AddressVerified: pgtype.Bool{Bool: true, Valid: true},
		CardCompany:     pgtype.Text{String: "Visa", Valid: true},
		City:            pgtype.Text{String: payload.City, Valid: true},
		ExpirationDate:  pgtype.Text{String: payload.ExpirationDate, Valid: true},
		FirstName:       pgtype.Text{String: payload.FirstName, Valid: true},
		LastName:        pgtype.Text{String: payload.LastName, Valid: true},
		LastFourDigits:  pgtype.Text{String: lastFour, Valid: true},
		FirstSixDigits:  pgtype.Text{String: firstSix, Valid: true},
		CardType:        pgtype.Text{String: "Credit", Valid: true},
		PullEnabled:     pgtype.Bool{Bool: true, Valid: true},
		PushEnabled:     pgtype.Bool{Bool: true, Valid: true},
		ReviewStatus:    pgtype.Text{String: "approved", Valid: true},
		Status:          pgtype.Text{String: "approved", Valid: true},
	})
	if err != nil {
		log.Error("failed to create card", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.UserCard{
		ID:              utils.BytesToUUID(card.ID.Bytes),
		AddressVerified: card.AddressVerified.Bool,
		CardCompany:     card.CardCompany.String,
		City:            card.City.String,
		ExpirationDate:  card.ExpirationDate.String,
		FirstName:       card.FirstName.String,
		LastName:        card.LastName.String,
		FirstSixDigits:  card.FirstSixDigits.String,
		LastFourDigits:  card.LastFourDigits.String,
		CardType:        card.CardType.String,
		PullEnabled:     card.PullEnabled.Bool,
		PushEnabled:     card.PushEnabled.Bool,
		Removed:         card.Removed.Bool,
		ReviewStatus:    card.ReviewStatus.String,
		State:           card.State.String,
		Status:          card.Status.String,
		StreetLine1:     card.StreetLine1.String,
		StreetLine2:     card.StreetLine2.String,
		ZipCode:         card.ZipCode.String,
		Created:         card.CreatedAt.Time.Format(time.RFC3339),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetUserCard(w http.ResponseWriter, r *http.Request) {
	accessToken := getBearerToken(r)
	parsedAccessToken, err := uuid.Parse(accessToken)
	if err != nil {
		log.Error("failed to parse access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token, err := s.q.GetAstraAccessToken(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedAccessToken), Valid: true})
	if err != nil {
		log.Error("failed to get access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	parsedCardID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to parse card id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	card, err := s.q.GetAstraUserCard(r.Context(), db.GetAstraUserCardParams{
		ID:     pgtype.UUID{Bytes: [16]byte(parsedCardID), Valid: true},
		UserID: token.UserID,
	})
	if err != nil {
		log.Error("failed to get card", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.UserCard{
		ID:              utils.BytesToUUID(card.ID.Bytes),
		AddressVerified: card.AddressVerified.Bool,
		CardCompany:     card.CardCompany.String,
		City:            card.City.String,
		ExpirationDate:  card.ExpirationDate.String,
		FirstName:       card.FirstName.String,
		LastName:        card.LastName.String,
		FirstSixDigits:  card.FirstSixDigits.String,
		LastFourDigits:  card.LastFourDigits.String,
		CardType:        card.CardType.String,
		PullEnabled:     card.PullEnabled.Bool,
		PushEnabled:     card.PushEnabled.Bool,
		Removed:         card.Removed.Bool,
		ReviewStatus:    card.ReviewStatus.String,
		State:           card.State.String,
		Status:          card.Status.String,
		StreetLine1:     card.StreetLine1.String,
		StreetLine2:     card.StreetLine2.String,
		ZipCode:         card.ZipCode.String,
		Created:         card.CreatedAt.Time.Format(time.RFC3339),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateUserAccount(w http.ResponseWriter, r *http.Request) {
	accessToken := getBearerToken(r)
	parsedAccessToken, err := uuid.Parse(accessToken)
	if err != nil {
		log.Error("failed to parse access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token, err := s.q.GetAstraAccessToken(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedAccessToken), Valid: true})
	if err != nil {
		log.Error("failed to get access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var payload external.CreateAccountArgs
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("failed to unmarhsal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	mask := payload.RoutingNumber
	if len(payload.RoutingNumber) > 4 {
		mask = payload.RoutingNumber[len(payload.RoutingNumber)-4:]
	}

	account, err := s.q.CreateAstraUserAccount(r.Context(), db.CreateAstraUserAccountParams{
		UserID:           token.UserID,
		Name:             pgtype.Text{String: payload.Name, Valid: true},
		OfficialName:     pgtype.Text{String: payload.Name, Valid: true},
		InstitutionName:  pgtype.Text{String: "Test Bank", Valid: true},
		Type:             pgtype.Text{String: string(payload.BankAccountType), Valid: true},
		Mask:             pgtype.Text{String: mask, Valid: true},
		ConnectionStatus: pgtype.Text{String: "Connected", Valid: true},
	})
	if err != nil {
		log.Error("failed to create account", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.UserAccount{
		ID:               utils.BytesToUUID(account.ID.Bytes),
		OfficialName:     account.OfficialName.String,
		Name:             account.Name.String,
		Mask:             account.Mask.String,
		InstitutionName:  account.InstitutionName.String,
		Type:             external.AccountType(account.Type.String),
		Subtype:          account.Subtype.String,
		ConnectionStatus: account.ConnectionStatus.String,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetUserAccount(w http.ResponseWriter, r *http.Request) {
	accessToken := getBearerToken(r)
	parsedAccessToken, err := uuid.Parse(accessToken)
	if err != nil {
		log.Error("failed to parse access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token, err := s.q.GetAstraAccessToken(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedAccessToken), Valid: true})
	if err != nil {
		log.Error("failed to get access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	parsedAccountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to parse account id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	account, err := s.q.GetAstraUserAccount(r.Context(), db.GetAstraUserAccountParams{
		UserID: token.UserID,
		ID:     pgtype.UUID{Bytes: [16]byte(parsedAccountID), Valid: true},
	})
	if err != nil {
		log.Error("failed to get account", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.UserAccount{
		ID:               utils.BytesToUUID(account.ID.Bytes),
		OfficialName:     account.OfficialName.String,
		Name:             account.Name.String,
		Mask:             account.Mask.String,
		InstitutionName:  account.InstitutionName.String,
		Type:             external.AccountType(account.Type.String),
		Subtype:          account.Subtype.String,
		ConnectionStatus: account.ConnectionStatus.String,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) AccountToCard(w http.ResponseWriter, r *http.Request) {
	accessToken := getBearerToken(r)
	parsedAccessToken, err := uuid.Parse(accessToken)
	if err != nil {
		log.Error("failed to parse access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token, err := s.q.GetAstraAccessToken(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedAccessToken), Valid: true})
	if err != nil {
		log.Error("failed to get access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var payload external.AccountToCardArgs
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("failed to unmarhsal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	parsedAccountID, err := uuid.Parse(payload.Account.ID)
	if err != nil {
		log.Error("failed to parse account id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	account, err := s.q.GetAstraAccount(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedAccountID), Valid: true})
	if err != nil {
		log.Error("failed to get account", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if account.UserID != token.UserID {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	parsedCardID, err := uuid.Parse(payload.Card.ID)
	if err != nil {
		log.Error("failed to parse card id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	parsedCardUserID, err := uuid.Parse(payload.Card.UserID)
	if err != nil {
		log.Error("failed to parse card user id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	card, err := s.q.GetAstraUserCard(r.Context(), db.GetAstraUserCardParams{
		ID:     pgtype.UUID{Bytes: [16]byte(parsedCardID), Valid: true},
		UserID: pgtype.UUID{Bytes: [16]byte(parsedCardUserID), Valid: true},
	})
	if err != nil {
		log.Error("failed to get card", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	trxID := uuid.New()
	trx, err := s.q.CreateAstraTransaction(r.Context(), db.CreateAstraTransactionParams{
		ID:                pgtype.UUID{Bytes: [16]byte(trxID), Valid: true},
		RoutineType:       pgtype.Text{String: "one-time", Valid: true},
		RoutineName:       pgtype.Text{String: "account-to-card", Valid: true},
		RoutineID:         pgtype.Text{String: trxID.String(), Valid: true},
		SourceID:          account.ID,
		DestinationID:     card.ID,
		DestinationUserID: card.UserID,
		Status:            pgtype.Text{String: "processed", Valid: true},
		PaymentType:       pgtype.Text{String: "debit", Valid: true},
		Initiated:         pgtype.Text{String: time.Now().Format(time.RFC3339), Valid: true},
		Amount:            pgtype.Float8{Float64: payload.Amount},
	})
	if err != nil {
		log.Error("failed to create account to card transaction", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.CardToAccountResp{
		ID: utils.BytesToUUID(trx.ID.Bytes),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) CardToAccount(w http.ResponseWriter, r *http.Request) {
	accessToken := getBearerToken(r)
	parsedAccessToken, err := uuid.Parse(accessToken)
	if err != nil {
		log.Error("failed to parse access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token, err := s.q.GetAstraAccessToken(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedAccessToken), Valid: true})
	if err != nil {
		log.Error("failed to get access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var payload external.CardToAccountArgs
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("failed to unmarhsal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	parsedCardID, err := uuid.Parse(payload.Card.ID)
	if err != nil {
		log.Error("failed to parse card id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	card, err := s.q.GetAstraCard(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedCardID), Valid: true})
	if err != nil {
		log.Error("failed to get card", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if card.UserID != token.UserID {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	parsedAccountID, err := uuid.Parse(payload.Account.ID)
	if err != nil {
		log.Error("failed to parse account id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	parsedAccountUserID, err := uuid.Parse(payload.Account.UserID)
	if err != nil {
		log.Error("failed to parse account user id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	account, err := s.q.GetAstraUserAccount(r.Context(), db.GetAstraUserAccountParams{
		ID:     pgtype.UUID{Bytes: [16]byte(parsedAccountID), Valid: true},
		UserID: pgtype.UUID{Bytes: [16]byte(parsedAccountUserID), Valid: true},
	})
	if err != nil {
		log.Error("failed to get account", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	trxID := uuid.New()
	trx, err := s.q.CreateAstraTransaction(r.Context(), db.CreateAstraTransactionParams{
		ID:                pgtype.UUID{Bytes: [16]byte(trxID), Valid: true},
		RoutineType:       pgtype.Text{String: "one-time", Valid: true},
		RoutineName:       pgtype.Text{String: "card-to-account", Valid: true},
		RoutineID:         pgtype.Text{String: trxID.String(), Valid: true},
		SourceID:          card.ID,
		DestinationID:     account.ID,
		DestinationUserID: account.UserID,
		Status:            pgtype.Text{String: "processed", Valid: true},
		PaymentType:       pgtype.Text{String: "debit", Valid: true},
		Initiated:         pgtype.Text{String: time.Now().Format(time.RFC3339), Valid: true},
		Amount:            pgtype.Float8{Float64: payload.Amount},
	})
	if err != nil {
		log.Error("failed to create card to account transaction", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.CardToAccountResp{
		ID: utils.BytesToUUID(trx.ID.Bytes),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetTransfer(w http.ResponseWriter, r *http.Request) {
	accessToken := getBearerToken(r)
	parsedAccessToken, err := uuid.Parse(accessToken)
	if err != nil {
		log.Error("failed to parse access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	_, err = s.q.GetAstraAccessToken(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedAccessToken), Valid: true})
	if err != nil {
		log.Error("failed to get access token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	parsedTrxID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to parse transfer id", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	trx, err := s.q.GetAstraTransaction(r.Context(), pgtype.UUID{Bytes: [16]byte(parsedTrxID), Valid: true})
	if err != nil {
		log.Error("failed to get transfer", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.Transaction{
		ID:                    utils.BytesToUUID(trx.ID.Bytes),
		RoutineType:           trx.RoutineType.String,
		RoutineName:           trx.RoutineName.String,
		RoutineID:             trx.RoutineID.String,
		Amount:                trx.Amount.Float64,
		AstraSettlementReason: trx.AstraSettlementReason.String,
		FailureReason:         trx.FailureReason.String,
		Status:                trx.Status.String,
		ClientCorrelationID:   trx.ClientCorrelationID.String,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("failed to return response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func getBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if bearerTokenRegexp.MatchString(authHeader) {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
