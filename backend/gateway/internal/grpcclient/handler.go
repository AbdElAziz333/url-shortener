package grpcclient

import (
	"net/http"
	"strconv"
	"time"

	"aziz.dev/gateway/internal/middleware"
	linkpb "aziz.dev/gateway/internal/pb/link"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// grpcToHTTP maps gRPC status codes to HTTP statuses; anything missing becomes a 500.
var grpcToHTTP = map[codes.Code]int{
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.NotFound:           http.StatusNotFound,
	codes.AlreadyExists:      http.StatusConflict,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.FailedPrecondition: http.StatusGone,
	codes.Unavailable:        http.StatusServiceUnavailable,
	codes.DeadlineExceeded:   http.StatusGatewayTimeout,
}

type Handler struct {
	clients *Clients
}

func NewHandler(clients *Clients) *Handler {
	return &Handler{clients: clients}
}

func (h *Handler) ListLinks(c *gin.Context) {
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

    resp, err := h.clients.Link.ListLinks(c.Request.Context(), &linkpb.ListLinksRequest{
        UserId: currentUserID(c),
        Limit:  int32(limit),
        Offset: int32(offset),
    })
    if err != nil {
        writeGRPCError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": linksToJSON(resp.GetLinks())})
}

func (h *Handler) GetLink(c *gin.Context) {
	resp, err := h.clients.Link.GetLink(c.Request.Context(), &linkpb.GetLinkRequest{
		UserId: currentUserID(c),
		Code: c.Param("code"),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": linkToJSON(resp.GetLink())})
}

func (h *Handler) CreateLink(c *gin.Context) {
	var body struct {
		OriginalURL string     `json:"original_url" binding:"required,url"`
		ExpiresAt   *time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := &linkpb.CreateLinkRequest{
		UserId:      currentUserID(c),
		OriginalUrl: body.OriginalURL,
	}
	if body.ExpiresAt != nil {
		req.ExpiresAt = timestamppb.New(*body.ExpiresAt)
	}

	resp, err := h.clients.Link.CreateLink(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": linkToJSON(resp.GetLink())})
}

func (h *Handler) UpdateExpiry(c *gin.Context) {
	var body struct {
		ExpiresAt time.Time `json:"expires_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.clients.Link.UpdateLinkExpiry(c.Request.Context(), &linkpb.UpdateLinkExpiryRequest{
		UserId:    currentUserID(c),
		Code:      c.Param("code"),
		ExpiresAt: timestamppb.New(body.ExpiresAt),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully updated expiry"})
}

func (h *Handler) DeleteLink(c *gin.Context) {
	_, err := h.clients.Link.DeleteLink(c.Request.Context(), &linkpb.DeleteLinkRequest{
		UserId: currentUserID(c),
		Code:   c.Param("code"),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully deleted link"})
}

// currentUserID returns the ID set by AccessTokenMiddleware, which has already
// validated it, so the routes using this must sit behind that middleware.
func currentUserID(c *gin.Context) string {
	return c.GetString(middleware.UserIDKey)
}

// writeGRPCError translates a gRPC error into an HTTP response. Server-side
// failures are logged and returned with a generic message so internals don't leak.
func writeGRPCError(c *gin.Context, err error) {
	st := status.Convert(err)

	httpStatus, ok := grpcToHTTP[st.Code()]
	if !ok {
		httpStatus = http.StatusInternalServerError
	}

	msg := st.Message()
	if httpStatus >= http.StatusInternalServerError {
		logrus.WithError(err).WithField("path", c.FullPath()).Error("gRPC call failed")
		msg = http.StatusText(httpStatus)
	}

	c.JSON(httpStatus, gin.H{"error": msg})
}

func linksToJSON(links []*linkpb.Link) []gin.H {
	out := make([]gin.H, 0, len(links))
	for _, link := range links {
		out = append(out, linkToJSON(link))
	}
	return out
}

func linkToJSON(link *linkpb.Link) gin.H {
	if link == nil {
		return gin.H{}
	}

	item := gin.H{
		"code":         link.GetCode(),
		"original_url": link.GetOriginalUrl(),
		"is_active":    link.GetIsActive(),
	}
	if link.ExpiresAt != nil {
		item["expires_at"] = link.GetExpiresAt().AsTime()
	}
	if link.CreatedAt != nil {
		item["created_at"] = link.GetCreatedAt().AsTime()
	}
	return item
}