package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jambotails/shipping-service/internal/services"
	"github.com/jambotails/shipping-service/pkg/response"
	"github.com/jambotails/shipping-service/pkg/validator"
)

// ShippingHandler handles shipping charge API requests.
type ShippingHandler struct {
	shippingSvc *services.ShippingService
}

// NewShippingHandler creates a new ShippingHandler.
func NewShippingHandler(ss *services.ShippingService) *ShippingHandler {
	return &ShippingHandler{shippingSvc: ss}
}

// CalculateChargeRequest is the JSON body for POST /api/v1/shipping-charge/calculate.
type CalculateChargeRequest struct {
	SellerID      int64  `json:"sellerId" validate:"required,gt=0"`
	ProductID     int64  `json:"productId" validate:"required,gt=0"`
	CustomerID    int64  `json:"customerId" validate:"required,gt=0"`
	DeliverySpeed string `json:"deliverySpeed" validate:"required"`
}

// GetCharge handles GET /api/v1/shipping-charge?warehouseId=&productId=&customerId=&deliverySpeed=
func (h *ShippingHandler) GetCharge(c *gin.Context) {
	warehouseIDStr := c.Query("warehouseId")
	productIDStr := c.Query("productId")
	customerIDStr := c.Query("customerId")
	deliverySpeed := c.Query("deliverySpeed")

	if warehouseIDStr == "" || productIDStr == "" || customerIDStr == "" || deliverySpeed == "" {
		c.JSON(http.StatusBadRequest,
			response.Error(c, http.StatusBadRequest, "warehouseId, productId, customerId, and deliverySpeed are required"),
		)
		return
	}

	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(http.StatusBadRequest, response.Error(c, http.StatusBadRequest, "warehouseId must be a positive integer"))
		return
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, response.Error(c, http.StatusBadRequest, "productId must be a positive integer"))
		return
	}

	customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
	if err != nil || customerID <= 0 {
		c.JSON(http.StatusBadRequest, response.Error(c, http.StatusBadRequest, "customerId must be a positive integer"))
		return
	}

	if err := validator.ValidateDeliverySpeed(deliverySpeed); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(c, http.StatusBadRequest, err.Error()))
		return
	}

	result, svcErr := h.shippingSvc.CalculateCharge(
		c.Request.Context(), warehouseID, customerID, productID, deliverySpeed,
	)
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, response.Error(c, http.StatusInternalServerError, svcErr.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(c, result))
}

// CalculateFull handles POST /api/v1/shipping-charge/calculate.
// It finds the nearest warehouse and then computes the shipping charge.
func (h *ShippingHandler) CalculateFull(c *gin.Context) {
	var req CalculateChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(c, http.StatusBadRequest, "invalid JSON body"))
		return
	}

	if err := validator.AppValidator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":   false,
			"requestId": c.GetString("requestId"),
			"error":     "validation failed",
			"fields":    validator.FormatValidationErrors(err),
		})
		return
	}

	if err := validator.ValidateDeliverySpeed(req.DeliverySpeed); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(c, http.StatusBadRequest, err.Error()))
		return
	}

	result, svcErr := h.shippingSvc.CalculateFull(
		c.Request.Context(), req.SellerID, req.CustomerID, req.ProductID, req.DeliverySpeed,
	)
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, response.Error(c, http.StatusInternalServerError, svcErr.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(c, result))
}
