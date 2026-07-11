// costeo_test.go — the engine must reproduce the EXACT numbers of the design
// prototype (Component class), which are also the ones shown in the handoff
// README: Choux $135.35 (with nested Mousseline $106.00), Buttercream $182.92,
// Pastel $50.26, Muffin $65.87, huevo $1.88/pza, harina $17.73/kg.
package services

import (
	"math"
	"testing"

	"idilica-backend-go/internal/models"
	"idilica-backend-go/internal/seeddata"
)

func buildCatalogo() *Catalogo {
	ingredientes, recetas := seeddata.BuildModels()
	return &Catalogo{Ingredientes: ingredientes, Recetas: recetas}
}

func almostEqual(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Errorf("%s: got %.4f, want %.4f", name, got, want)
	}
}

func TestCostoIngredienteContraPrototipo(t *testing.T) {
	c := buildCatalogo()

	// harina: 780 / 44 = 17.7273 (sin merma)
	almostEqual(t, "harina/kg", CostoIngrediente(c.Ingredientes["harina"]), 17.7273, 0.001)
	// huevo: 49.50 / 30 / (1 − 0.12) = 1.875 → "$1.88/pza"
	almostEqual(t, "huevo/pza", CostoIngrediente(c.Ingredientes["huevo"]), 1.875, 0.0001)
	// piña: 27.50 / 1 / 0.75 = 36.6667 → "$36.67 ya con desperdicio"
	almostEqual(t, "pina/kg", CostoIngrediente(c.Ingredientes["pina"]), 36.6667, 0.001)
	// chocolate: 961.25 / 2.5 = 384.50
	almostEqual(t, "chocolate/kg", CostoIngrediente(c.Ingredientes["chocolate"]), 384.50, 0.0001)
}

func TestCostoRecetaContraPrototipo(t *testing.T) {
	c := buildCatalogo()

	almostEqual(t, "mousseline", c.CostoReceta("mousseline"), 106.00, 0.01)
	almostEqual(t, "buttercream", c.CostoReceta("buttercream"), 182.92, 0.01)
	almostEqual(t, "choux (con subreceta)", c.CostoReceta("choux"), 135.35, 0.01)
	almostEqual(t, "pastel", c.CostoReceta("pastel"), 50.26, 0.01)
	almostEqual(t, "muffin", c.CostoReceta("muffin"), 65.87, 0.01)
}

func TestFoodCostYNivel(t *testing.T) {
	c := buildCatalogo()

	// choux: 135.35 / 250 = 54.1% → rojo con objetivo 30
	pct := c.FoodCostPct("choux")
	if pct == nil {
		t.Fatal("choux should have a food cost")
	}
	almostEqual(t, "choux fc%", *pct, 54.14, 0.05)
	if nivel := Nivel(pct, seeddata.ObjetivoFoodCost); nivel != "rojo" {
		t.Errorf("choux nivel: got %s, want rojo", nivel)
	}

	// pastel: 50.26 / 450 = 11.2% → verde
	if nivel := Nivel(c.FoodCostPct("pastel"), seeddata.ObjetivoFoodCost); nivel != "verde" {
		t.Errorf("pastel nivel: got %s, want verde", nivel)
	}

	// muffin: 65.87 / 190 = 34.7% → ámbar (objetivo 30, límite 40)
	if nivel := Nivel(c.FoodCostPct("muffin"), seeddata.ObjetivoFoodCost); nivel != "ambar" {
		t.Errorf("muffin nivel: got %s, want ambar", nivel)
	}

	// subreceta sin precio → gris
	if nivel := Nivel(c.FoodCostPct("mousseline"), seeddata.ObjetivoFoodCost); nivel != "gris" {
		t.Errorf("mousseline nivel: got %s, want gris", nivel)
	}
}

func TestUsaIngredienteRecursivo(t *testing.T) {
	c := buildCatalogo()

	// choux usa almendra SOLO a través de la subreceta mousseline
	if !c.UsaIngrediente("choux", "almendra") {
		t.Error("choux should use almendra via mousseline")
	}
	if c.UsaIngrediente("croissant", "almendra") {
		t.Error("croissant should not use almendra")
	}
}

func TestDeteccionDeCiclos(t *testing.T) {
	c := buildCatalogo()

	// choux → mousseline ya existe; agregar choux DENTRO de mousseline sería un ciclo
	if !c.UsaReceta("choux", "mousseline") {
		t.Error("choux reaches mousseline")
	}
	if c.UsaReceta("mousseline", "choux") {
		t.Error("mousseline must not reach choux (would be a cycle)")
	}

	// Y aunque los datos vinieran corruptos con un ciclo, el motor no se cuelga
	chouxID := "choux"
	c.Recetas["mousseline"].Lineas = append(c.Recetas["mousseline"].Lineas, models.RecetaLinea{
		RecetaID: "mousseline", SubRecetaID: &chouxID, Cantidad: 0.1,
	})
	_ = c.CostoReceta("choux") // must terminate
}

func TestGatherNeedsRecursivo(t *testing.T) {
	c := buildCatalogo()

	// Choux usa mousseline (0.95 kg / rinde 0.95 → mult 1): las necesidades
	// suman lo directo + lo de la subreceta.
	needs := c.GatherNeeds("choux")
	almostEqual(t, "harina", needs["harina"], 0.25, 1e-9)
	almostEqual(t, "mantequilla", needs["mantequilla"], 0.1+0.25, 1e-9) // directa + sub
	almostEqual(t, "huevo", needs["huevo"], 4+4, 1e-9)
	almostEqual(t, "leche", needs["leche"], 0.25+0.2, 1e-9)
	almostEqual(t, "azucar", needs["azucar"], 0.0315+0.12, 1e-9)
	almostEqual(t, "almendra (solo vía sub)", needs["almendra"], 0.2, 1e-9)
}

func TestTasaOperacion(t *testing.T) {
	cocina := &models.Cocina{
		GastoSueldos: seeddata.Opex.Sueldos, GastoGas: seeddata.Opex.Gas,
		GastoLuz: seeddata.Opex.Luz, GastoEquipo: seeddata.Opex.Equipo,
		ComprasIngredientesMes: seeddata.Opex.Compras,
	}
	// (12000+1800+1400+900)/38000 = 0.4237 → "Por cada $100 … $42.37"
	almostEqual(t, "tasa operación", TasaOperacion(cocina), 0.4237, 0.0001)

	cocina.ComprasIngredientesMes = 0
	if TasaOperacion(cocina) != 0 {
		t.Error("tasa must be 0 when compras is 0")
	}
}
