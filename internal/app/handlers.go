package app

import (
	"errors"
	"io"
	"net/http"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/service"
	"github.com/Pizhlo/yandex-shortener/internal/app/session"
	storage "github.com/Pizhlo/yandex-shortener/storage/db"
	errs "github.com/Pizhlo/yandex-shortener/storage/errors"
	"github.com/Pizhlo/yandex-shortener/storage/model"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	Service      *service.Service
	Logger       log.Logger
	FlagBaseAddr string
}

func ReceiveURL(handler Handler, w http.ResponseWriter, r *http.Request) {
	handler.Logger.Sugar.Debug("ReceiveUrl")

	// сократить ссылку
	// записать в базу

	j, err := io.ReadAll(r.Body)
	if err != nil {
		handler.Logger.Sugar.Debug("ReceiveUrl ReadAll err = ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	statusCode := http.StatusCreated
	shortURL := util.Shorten(string(j))

	ctx := r.Context()

	var userID any
	cookie, err := r.Cookie("token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			userID = ctx.Value("userID")
			handler.Logger.Sugar.Debug("ReceiveUrl userID = ", userID)
		} else {
			handler.Logger.Sugar.Debug("ReceiveUrl Cookie err = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		userID, err = session.GetUserID(cookie.Value)
		if err != nil {
			handler.Logger.Sugar.Debug("ReceiveUrl GetUserID err = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	linkModel, err := model.MakeLinkModel("", userID.(uuid.UUID), shortURL, string(j))
	if err != nil {
		handler.Logger.Sugar.Debug("ReceiveUrl MakeLinkModel err = ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := handler.Service.Storage.Save(ctx, linkModel); err != nil {
		handler.Logger.Sugar.Debug("ReceiveUrl SaveLink err = ", err)
		if err.Error() == uniqueViolation {
			statusCode = http.StatusConflict
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	handler.Logger.Sugar.Debug("ReceiveUrl code = ", statusCode)

	path, err := util.MakeURL(handler.FlagBaseAddr, shortURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setHeader(w, "Content-Type", "text/plain", statusCode)

	w.Write([]byte(path))
}

func GetURL(handler Handler, w http.ResponseWriter, r *http.Request) {
	handler.Logger.Sugar.Debug("GetUrl")

	// проверить наличие ссылки в базе
	// выдать ссылку

	id := chi.URLParam(r, "id")

	ctx := r.Context()

	val, err := handler.Service.Storage.Get(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	setHeader(w, "Location", val, http.StatusTemporaryRedirect)
}

func Ping(w http.ResponseWriter, r *http.Request, db *storage.URLStorage) {
	// ping

	err := db.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func setHeader(w http.ResponseWriter, header string, val string, statusCode int) {
	w.Header().Set(header, val)
	w.WriteHeader(statusCode)
}
