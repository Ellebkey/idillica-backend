// costeo.go — the costing engine, ported from the `Component` class of the
// design handoff (the executable spec). Pure functions over in-memory data;
// costs are NEVER persisted, always derived.
package services

import (
	"idilica-backend-go/internal/models"
)

// ProductoActivo returns the active purchase product of an ingredient
// (falls back to the first one so a half-configured ingredient never panics).
func ProductoActivo(ing *models.Ingrediente) *models.ProductoCompra {
	if ing == nil || len(ing.Productos) == 0 {
		return nil
	}
	for i := range ing.Productos {
		if ing.Productos[i].Activo {
			return &ing.Productos[i]
		}
	}
	return &ing.Productos[0]
}

// CostoIngrediente ≈ ingCost(): precio del producto activo / cantidad de la
// presentación / (1 − merma/100) → "costo por unidad base, ya con desperdicio".
func CostoIngrediente(ing *models.Ingrediente) float64 {
	prod := ProductoActivo(ing)
	if prod == nil || prod.Cantidad <= 0 {
		return 0
	}
	factor := 1 - ing.MermaPct/100
	if factor <= 0 {
		factor = 1
	}
	return prod.Precio / prod.Cantidad / factor
}

// Catalogo is the in-memory graph the engine walks (maps by ID).
type Catalogo struct {
	Ingredientes map[string]*models.Ingrediente
	Recetas      map[string]*models.Receta
}

// CostoReceta ≈ recCost(): Σ líneas; línea ingrediente = cantidad × costo del
// ingrediente; línea subreceta = (cantidadKg / rendimientoKg) × costo de la
// subreceta, recursivo. `visited` corta ciclos defensivamente (el service los
// rechaza al guardar, pero el motor nunca debe colgarse con datos corruptos).
func (c *Catalogo) CostoReceta(recetaID string) float64 {
	return c.costoReceta(recetaID, map[string]bool{})
}

func (c *Catalogo) costoReceta(recetaID string, visited map[string]bool) float64 {
	receta, ok := c.Recetas[recetaID]
	if !ok || visited[recetaID] {
		return 0
	}
	visited[recetaID] = true
	defer delete(visited, recetaID)

	total := 0.0
	for _, linea := range receta.Lineas {
		switch {
		case linea.IngredienteID != nil:
			total += linea.Cantidad * CostoIngrediente(c.Ingredientes[*linea.IngredienteID])
		case linea.SubRecetaID != nil:
			sub, ok := c.Recetas[*linea.SubRecetaID]
			if !ok || sub.RendimientoKg <= 0 {
				continue
			}
			total += (linea.Cantidad / sub.RendimientoKg) * c.costoReceta(*linea.SubRecetaID, visited)
		}
	}
	return total
}

// UsaIngrediente ≈ usesIng(): true if the recipe uses the ingredient directly
// or through any nested sub-recipe.
func (c *Catalogo) UsaIngrediente(recetaID, ingredienteID string) bool {
	return c.usaIngrediente(recetaID, ingredienteID, map[string]bool{})
}

func (c *Catalogo) usaIngrediente(recetaID, ingredienteID string, visited map[string]bool) bool {
	receta, ok := c.Recetas[recetaID]
	if !ok || visited[recetaID] {
		return false
	}
	visited[recetaID] = true

	for _, linea := range receta.Lineas {
		if linea.IngredienteID != nil && *linea.IngredienteID == ingredienteID {
			return true
		}
		if linea.SubRecetaID != nil && c.usaIngrediente(*linea.SubRecetaID, ingredienteID, visited) {
			return true
		}
	}
	return false
}

// UsaReceta reports whether recetaID reaches targetID through its sub-recipe
// lines — the cycle check: adding `target` as a line of `receta` is illegal
// when UsaReceta(target, receta) is already true (or target == receta).
func (c *Catalogo) UsaReceta(recetaID, targetID string) bool {
	return c.usaReceta(recetaID, targetID, map[string]bool{})
}

func (c *Catalogo) usaReceta(recetaID, targetID string, visited map[string]bool) bool {
	if recetaID == targetID {
		return true
	}
	receta, ok := c.Recetas[recetaID]
	if !ok || visited[recetaID] {
		return false
	}
	visited[recetaID] = true

	for _, linea := range receta.Lineas {
		if linea.SubRecetaID != nil && c.usaReceta(*linea.SubRecetaID, targetID, visited) {
			return true
		}
	}
	return false
}

// FoodCostPct ≈ fcPct(): nil when the recipe has no sale price (subrecetas).
func (c *Catalogo) FoodCostPct(recetaID string) *float64 {
	receta, ok := c.Recetas[recetaID]
	if !ok || receta.PrecioVenta == nil || *receta.PrecioVenta <= 0 {
		return nil
	}
	pct := c.CostoReceta(recetaID) / *receta.PrecioVenta * 100
	return &pct
}

// Nivel ≈ level(): traffic-light level for a food-cost %, given the target
// (objetivo as whole percent, e.g. 30). ≤obj verde, ≤obj+10 ámbar, else rojo.
func Nivel(pct *float64, objetivo float64) string {
	switch {
	case pct == nil:
		return "gris"
	case *pct <= objetivo:
		return "verde"
	case *pct <= objetivo+10:
		return "ambar"
	default:
		return "rojo"
	}
}

// TasaOperacion ≈ opexRate(): Σ gastos mensuales / compras de ingredientes.
func TasaOperacion(cocina *models.Cocina) float64 {
	if cocina == nil || cocina.ComprasIngredientesMes <= 0 {
		return 0
	}
	gastos := cocina.GastoSueldos + cocina.GastoGas + cocina.GastoLuz + cocina.GastoEquipo
	return gastos / cocina.ComprasIngredientesMes
}
