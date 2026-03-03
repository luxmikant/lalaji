package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jambotails/shipping-service/internal/services"
	apperrors "github.com/jambotails/shipping-service/pkg/errors"
	"github.com/jambotails/shipping-service/pkg/response"
)

// WarehouseHandler handles nearest-warehouse API requests.
type WarehouseHandler struct {
	warehouseSvc *services.WarehouseService
}

// NewWarehouseHandler creates a new WarehouseHandler.
func NewWarehouseHandler(ws *services.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{warehouseSvc: ws}
}

// FindNearest handles GET /api/v1/warehouse/nearest?sellerId=1&productId=2
func (h *WarehouseHandler) FindNearest(c *gin.Context) {
	sellerIDStr := c.Query("sellerId")
	productIDStr := c.Query("productId")

	if sellerIDStr == "" || productIDStr == "" {
		c.JSON(http.StatusBadRequest,
			response.Error(c, http.StatusBadRequest, "sellerId and productId are required query parameters"),
		)
		return
	}

	sellerID, err := strconv.ParseInt(sellerIDStr, 10, 64)
	if err != nil || sellerID <= 0 {
		c.JSON(http.StatusBadRequest,
			response.Error(c, http.StatusBadRequest, "sellerId must be a positive integer"),
		)
		return
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest,
			response.Error(c, http.StatusBadRequest, "productId must be a positive integer"),
		)
		return
	}

	result, svcErr := h.warehouseSvc.FindNearest(c.Request.Context(), sellerID, productID)
	if svcErr != nil {
		// Unwrap typed AppError for the correct HTTP status code (404, 503, etc.)
		if appErr, ok := apperrors.AsAppError(svcErr); ok {
			c.JSON(appErr.HTTPStatus, response.Error(c, appErr.HTTPStatus, appErr.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(c, http.StatusInternalServerError, svcErr.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(c, result))
}
