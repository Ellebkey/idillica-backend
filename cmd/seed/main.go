// cmd/seed — populates a user's cocina with the demo dataset of the design
// handoff (15 ingredients, 8 recipes with sub-recipes, opex defaults).
//
//	go run ./cmd/seed -email prueba@idilica.app          # only if empty
//	go run ./cmd/seed -email prueba@idilica.app -force   # wipe & reseed
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"

	"idilica-backend-go/internal/config"
	"idilica-backend-go/internal/models"
	"idilica-backend-go/internal/seeddata"
)

func main() {
	email := flag.String("email", "", "email (username) del usuario dueño de la cocina a sembrar")
	force := flag.Bool("force", false, "borrar el catálogo existente y volver a sembrar")
	flag.Parse()

	if *email == "" {
		fmt.Fprintln(os.Stderr, "uso: go run ./cmd/seed -email <usuario> [-force]")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logger := config.NewLogger(cfg)
	db, err := config.NewDatabase(cfg, logger)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := run(db, *email, *force); err != nil {
		fmt.Fprintln(os.Stderr, "seed error:", err)
		os.Exit(1)
	}
	fmt.Println("✓ Catálogo demo sembrado")
}

func run(db *gorm.DB, email string, force bool) error {
	var user models.User
	if err := db.First(&user, "username = ?", email).Error; err != nil {
		return fmt.Errorf("usuario %q no encontrado (regístralo primero): %w", email, err)
	}

	var member models.CocinaMember
	if err := db.Preload("Cocina").First(&member, "user_id = ? AND rol = ?", user.ID, models.RolOwner).Error; err != nil {
		return fmt.Errorf("el usuario no tiene cocina propia: %w", err)
	}
	cocinaID := member.CocinaID

	var existentes int64
	db.Model(&models.Ingrediente{}).Where("cocina_id = ?", cocinaID).Count(&existentes)
	if existentes > 0 && !force {
		return fmt.Errorf("la cocina ya tiene %d ingredientes; usa -force para reemplazar", existentes)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if force {
			if err := wipe(tx, cocinaID); err != nil {
				return err
			}
		}

		// --- Opex + objetivo del handoff en la cocina ---
		err := tx.Model(&models.Cocina{}).Where("id = ?", cocinaID).Updates(map[string]any{
			"name":                     "Idílica Panadería Gourmet",
			"food_cost_objetivo":       seeddata.ObjetivoFoodCost / 100.0,
			"gasto_sueldos":            seeddata.Opex.Sueldos,
			"gasto_gas":                seeddata.Opex.Gas,
			"gasto_luz":                seeddata.Opex.Luz,
			"gasto_equipo":             seeddata.Opex.Equipo,
			"compras_ingredientes_mes": seeddata.Opex.Compras,
		}).Error
		if err != nil {
			return err
		}

		// --- Ingredientes + productos + historial ---
		ingIDs := map[string]string{} // slug → uuid
		now := time.Now()
		for _, seed := range seeddata.Ingredientes {
			ing := models.Ingrediente{
				CocinaID: cocinaID, Nombre: seed.Nombre, UnidadBase: seed.Unidad,
				MermaPct: seed.MermaPct, MermaOrigen: seed.MermaOrigen,
			}
			if stock, ok := seeddata.Stocks[seed.Slug]; ok {
				ing.Existencia = stock.Existencia
				ing.Minimo = stock.Minimo
				caduca := now.AddDate(0, 0, stock.CaducaDias)
				ing.CaducaAt = &caduca
			}
			if escalado, ok := seeddata.Escalados[seed.Slug]; ok {
				ing.Escalado = escalado
			}
			if err := tx.Create(&ing).Error; err != nil {
				return err
			}
			ingIDs[seed.Slug] = ing.ID

			fechaPrecio := now.AddDate(0, 0, -seed.DiasPrecio)
			for i, p := range seed.Productos {
				producto := models.ProductoCompra{
					IngredienteID: ing.ID, Marca: p.Marca, Presentacion: p.Presentacion,
					Cantidad: p.Cantidad, Precio: p.Precio, Proveedor: p.Proveedor,
					Activo: i == seed.ActivoIdx, Orden: i, PrecioActualizadoAt: fechaPrecio,
				}
				if err := tx.Create(&producto).Error; err != nil {
					return err
				}

				// Historial: una curva creíble hacia el precio actual, para que
				// el sparkline y la tendencia del detalle tengan datos reales.
				precios := historiaPara(p.Precio)
				for j, precio := range precios {
					meses := len(precios) - 1 - j
					h := models.HistorialPrecio{
						ProductoID: producto.ID, Precio: precio,
						Fecha: fechaPrecio.AddDate(0, -2*meses, 0),
					}
					if err := tx.Create(&h).Error; err != nil {
						return err
					}
				}
			}

			// La piña trae su medición de báscula (merma 25% "medido")
			if seed.Slug == "pina" {
				medicion := models.MedicionMerma{
					IngredienteID: ing.ID, PesoEntero: 1.2, PesoLimpio: 0.9,
					Aprovechado: 0, PctResultante: 25,
				}
				if err := tx.Create(&medicion).Error; err != nil {
					return err
				}
			}
		}

		// --- Equipo ---
		for _, eq := range seeddata.Equipos {
			h := models.Herramienta{CocinaID: cocinaID, Nombre: eq.Nombre, Detalle: eq.Detalle, Estado: eq.Estado}
			if err := tx.Create(&h).Error; err != nil {
				return err
			}
		}

		// --- Recetas (subrecetas primero; seeddata ya viene ordenado) ---
		recIDs := map[string]string{}
		for _, seed := range seeddata.Recetas {
			receta := models.Receta{
				CocinaID: cocinaID, Nombre: seed.Nombre, Categoria: seed.Categoria,
				Porciones: seed.Porciones, Etiqueta: seed.Etiqueta, EtiquetaSingular: seed.EtiquetaSingular,
				RendimientoKg: seed.RendimientoKg, PrecioVenta: seed.PrecioVenta,
				IvaPct: 16, EsSubreceta: seed.EsSubreceta,
				Alergenos: seed.Alergenos, Pasos: seed.Pasos, Fotos: models.StringSlice{},
			}
			if err := tx.Create(&receta).Error; err != nil {
				return err
			}
			recIDs[seed.Slug] = receta.ID

			for i, l := range seed.Lineas {
				linea := models.RecetaLinea{RecetaID: receta.ID, Orden: i, Cantidad: l.Cantidad}
				if l.IngredienteSlug != "" {
					id := ingIDs[l.IngredienteSlug]
					linea.IngredienteID = &id
				} else {
					id, ok := recIDs[l.RecetaSlug]
					if !ok {
						return fmt.Errorf("subreceta %q no sembrada aún (orden en seeddata)", l.RecetaSlug)
					}
					linea.SubRecetaID = &id
				}
				if err := tx.Create(&linea).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// historiaPara builds 4 plausible history points ending at the current price
// (the handoff's sparkline shows a rising trend; mantequilla's last jump is
// the canonical 96.50 → 108 example).
func historiaPara(actual float64) []float64 {
	return []float64{
		round2(actual * 0.86),
		round2(actual * 0.90),
		round2(actual * 0.894), // 96.50 para mantequilla (108 × 0.894)
		actual,
	}
}

func round2(v float64) float64 { return float64(int(v*100+0.5)) / 100 }

func wipe(tx *gorm.DB, cocinaID string) error {
	// Orden inverso de dependencias
	if err := tx.Exec(`DELETE FROM receta_linea WHERE receta_id IN (SELECT id FROM receta WHERE cocina_id = ?)`, cocinaID).Error; err != nil {
		return err
	}
	if err := tx.Where("cocina_id = ?", cocinaID).Delete(&models.Receta{}).Error; err != nil {
		return err
	}
	if err := tx.Exec(`DELETE FROM historial_precio WHERE producto_id IN
		(SELECT p.id FROM producto_compra p JOIN ingrediente i ON i.id = p.ingrediente_id WHERE i.cocina_id = ?)`, cocinaID).Error; err != nil {
		return err
	}
	if err := tx.Exec(`DELETE FROM producto_compra WHERE ingrediente_id IN (SELECT id FROM ingrediente WHERE cocina_id = ?)`, cocinaID).Error; err != nil {
		return err
	}
	if err := tx.Exec(`DELETE FROM medicion_merma WHERE ingrediente_id IN (SELECT id FROM ingrediente WHERE cocina_id = ?)`, cocinaID).Error; err != nil {
		return err
	}
	if err := tx.Where("cocina_id = ?", cocinaID).Delete(&models.Herramienta{}).Error; err != nil {
		return err
	}
	return tx.Where("cocina_id = ?", cocinaID).Delete(&models.Ingrediente{}).Error
}
