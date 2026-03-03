package models_test

import (
	"testing"

	"github.com/jambotails/shipping-service/internal/models"
)

func TestBillableWeightKg_ActualGreater(t *testing.T) {
	p := models.Product{
		ActualWeightKg: 10,
		LengthCm:       20,
		WidthCm:        20,
		HeightCm:       20,
		// Volumetric = 20*20*20/5000 = 1.6 kg
	}
	bw := p.BillableWeightKg()
	if bw != 10 {
		t.Errorf("expected 10 (actual), got %.2f", bw)
	}
}

func TestBillableWeightKg_VolumetricGreater(t *testing.T) {
	p := models.Product{
		ActualWeightKg: 1,
		LengthCm:       100,
		WidthCm:        100,
		HeightCm:       100,
		// Volumetric = 100*100*100/5000 = 200 kg
	}
	bw := p.BillableWeightKg()
	expected := (100.0 * 100.0 * 100.0) / 5000.0
	if bw != expected {
		t.Errorf("expected %.2f (volumetric), got %.2f", expected, bw)
	}
}

func TestBillableWeightKg_Equal(t *testing.T) {
	p := models.Product{
		ActualWeightKg: 5,
		LengthCm:       50,
		WidthCm:        50,
		HeightCm:       10,
		// Volumetric = 50*50*10/5000 = 5 kg
	}
	bw := p.BillableWeightKg()
	if bw != 5 {
		t.Errorf("expected 5 (equal), got %.2f", bw)
	}
}

func TestBillableWeightKg_SmallProduct(t *testing.T) {
	p := models.Product{
		ActualWeightKg: 0.5,
		LengthCm:       10,
		WidthCm:        5,
		HeightCm:       2,
		// Volumetric = 10*5*2/5000 = 0.02 kg
	}
	bw := p.BillableWeightKg()
	if bw != 0.5 {
		t.Errorf("expected 0.5, got %.4f", bw)
	}
}

func TestBillableWeightKg_LargeBox(t *testing.T) {
	p := models.Product{
		ActualWeightKg: 2,
		LengthCm:       200,
		WidthCm:        100,
		HeightCm:       50,
		// Volumetric = 200*100*50/5000 = 200 kg
	}
	bw := p.BillableWeightKg()
	if bw != 200 {
		t.Errorf("expected 200 (volumetric), got %.2f", bw)
	}
}
