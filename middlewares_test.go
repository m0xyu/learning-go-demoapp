package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// 表示確認のため、ログ出力をバッファへ
	buf := new(bytes.Buffer)
	testApp.infoLog = log.New(buf, "", 0)

	handler := testApp.logger(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// ログにリクエストの情報が含まれていることを確認
	assert.Contains(t, buf.String(), "HTTP/1.1 GET /test")
}

func TestRecover(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := testApp.recover(panicHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Internal Server Error\n", w.Body.String())

	assert.Equal(t, "close", w.Header().Get("Connection"))
}

func TestRequireAuth(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("protected"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(contextWithAuth(req.Context(), true))
	w := httptest.NewRecorder()

	handler := testApp.requireAuth(testHandler)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "protected", w.Body.String())
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
}

func TestRequireAuth_NoAuthenticated(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("protected"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler := testApp.requireAuth(testHandler)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login?redirectTo=/test", w.Header().Get("Location"))
}

func TestAuthenticate_ValidSession(t *testing.T) {
	defer cleanupTestData(t)
	userID, err := testApp.userRepo.CreateUser(
		"session user",
		"session@test.com",
		"passwordgood",
		"avatar",
	)
	assert.NoError(t, err)
	assert.Greater(t, userID, 0)

	// セッションにユーザー情報をセットするためのハンドラー
	setupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testApp.session.Put(r, loggedInUserKey, "session@test.com")
		w.WriteHeader(http.StatusOK)
	})

	// 認証後のハンドラー
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, testApp.isAuthenticated(r))
		user := testApp.getUserFromContext(r.Context())
		assert.Equal(t, "session@test.com", user.Email)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// セッションをセットするためのリクエスト
	setUpChain := testApp.session.Enable(testApp.authenticate(setupHandler))
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	setUpChain.ServeHTTP(w1, req1)

	// 認証後のリクエスト（セッションを引き継ぐ）
	testChain := testApp.session.Enable(testApp.authenticate(testHandler))
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	if cookies := w1.Result().Cookies(); len(cookies) > 0 {
		for _, cookie := range cookies {
			req2.AddCookie(cookie)
		}
	}

	w2 := httptest.NewRecorder()
	testChain.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "OK", w2.Body.String())
}

func TestAuthenticate_RemoveSession(t *testing.T) {
	defer cleanupTestData(t)
	user, err := testApp.userRepo.CreateUser(
		"session user",
		"session@test.com",
		"passwordgood",
		"avatar",
	)
	assert.NoError(t, err)
	assert.Greater(t, user, 0)

	// セッションにユーザー情報をセットするためのハンドラー
	setupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testApp.session.Put(r, loggedInUserKey, "non-existent@test.com")
		w.WriteHeader(http.StatusOK)
	})

	// 認証後のハンドラー
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.False(t, testApp.isAuthenticated(r))
		user := r.Context().Value(contextUserKey)
		assert.Nil(t, user)
		assert.Equal(t, "", testApp.session.GetString(r, loggedInUserKey))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// セッションをセットするためのリクエスト
	setUpChain := testApp.session.Enable(testApp.authenticate(setupHandler))
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	setUpChain.ServeHTTP(w1, req1)

	// 認証後のリクエスト（セッションを引き継ぐ）
	testChain := testApp.session.Enable(testApp.authenticate(testHandler))
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	if cookies := w1.Result().Cookies(); len(cookies) > 0 {
		for _, cookie := range cookies {
			req2.AddCookie(cookie)
		}
	}

	w2 := httptest.NewRecorder()
	testChain.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "OK", w2.Body.String())
}

func contextWithAuth(ctx context.Context, isAuth interface{}) context.Context {
	return context.WithValue(ctx, contextAuthKey, isAuth)
}
