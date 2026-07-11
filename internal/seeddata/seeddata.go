// Package seeddata is the demo dataset of the design handoff (the `Component`
// class of Idilica Costeo.dc.html), verbatim. It is used by cmd/seed to
// populate a cocina AND by the costing-engine tests, so the tests guarantee
// the engine reproduces the prototype's numbers exactly.
package seeddata

import "idilica-backend-go/internal/models"

type Producto struct {
	Marca        string
	Presentacion string
	Cantidad     float64 // contenido en unidad base
	Precio       float64
	Proveedor    string
}

type Ingrediente struct {
	Slug        string
	Nombre      string
	Unidad      string
	MermaPct    float64
	MermaOrigen string
	ActivoIdx   int
	DiasPrecio  int // días desde la última actualización de precio
	Productos   []Producto
}

type Linea struct {
	IngredienteSlug string // XOR con RecetaSlug
	RecetaSlug      string
	Cantidad        float64
}

type Receta struct {
	Slug             string
	Nombre           string
	Categoria        string
	Porciones        int
	Etiqueta         string
	EtiquetaSingular string
	RendimientoKg    float64
	PrecioVenta      *float64 // nil = subreceta sin precio
	EsSubreceta      bool
	Alergenos        []string
	Pasos            []string
	Lineas           []Linea
}

func precio(v float64) *float64 { return &v }

var Alergenos = []string{
	"Gluten", "Crustáceos", "Huevo", "Pescado", "Cacahuates", "Lácteos", "Apio",
	"Mostaza", "Sulfitos", "Sésamo", "Moluscos", "Soya", "Frutos secos", "Altramuz",
}

var Categorias = []string{"Pasteles", "Pastas", "Muffins", "Cupcakes", "Panes", "Complementos"}

// Opex defaults del handoff → tasa ≈ 42%.
var Opex = struct {
	Sueldos, Gas, Luz, Equipo, Compras float64
}{12000, 1800, 1400, 900, 38000}

// Stock — inventario semilla del handoff: existencia, mínimo (dispara "queda
// poco") y días para caducar (el seeder los convierte a fecha).
type Stock struct {
	Existencia, Minimo float64
	CaducaDias         int
}

var Stocks = map[string]Stock{
	"harina":      {26, 10, 180},
	"mantequilla": {4, 3, 24},
	"huevo":       {42, 30, 12},
	"azucar":      {8, 5, 365},
	"chocolate":   {2.4, 1, 200},
	"pina":        {3, 2, 6},
	"leche":       {6, 4, 10},
	"crema":       {1.5, 2, 8},
	"fresa":       {1.2, 2, 3},
	"frambuesa":   {0.8, 1, 60},
	"limon":       {2, 1, 15},
	"almendra":    {1.5, 1, 90},
	"vainilla":    {0.4, 0.25, 300},
	"canela":      {0.3, 0.2, 400},
	"levadura":    {0.6, 0.25, 45},
}

// Escalado no lineal del catálogo semilla (el resto queda "normal"):
// leudantes al 75% en lotes grandes, sazón con factor^0.7.
var Escalados = map[string]string{
	"levadura": "leudante",
	"canela":   "sazon",
	"vainilla": "sazon",
}

// Equipo semilla del handoff.
type Equipo struct {
	Nombre, Detalle, Estado string
}

var Equipos = []Equipo{
	{"Batidora planetaria 5 L", "KitchenAid · comprada en 2023", "Buen estado"},
	{"Horno de piso", "San-Son · servicio programado en agosto", "Servicio pronto"},
	{"Moldes redondos 22 cm", "6 piezas · aluminio", "Buen estado"},
	{"Charolas para muffin", "4 piezas · 12 cavidades · 1 con óxido", "Reponer 1"},
	{"Mangas y duyas", "12 duyas · 2 mangas reutilizables", "Buen estado"},
	{"Báscula digital", "Rhino 5 kg · precisión 1 g", "Buen estado"},
}

const ObjetivoFoodCost = 30 // % entero

var Ingredientes = []Ingrediente{
	{Slug: "harina", Nombre: "Harina de trigo", Unidad: "kg", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 12, Productos: []Producto{
		{Marca: "Harinera Elizondo", Presentacion: "Bulto 44 kg", Cantidad: 44, Precio: 780, Proveedor: "Abarrotes La Central"},
		{Marca: "Selecta", Presentacion: "Bolsa 1 kg", Cantidad: 1, Precio: 28, Proveedor: "Súper del barrio"},
	}},
	{Slug: "mantequilla", Nombre: "Mantequilla", Unidad: "kg", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 2, Productos: []Producto{
		{Marca: "Gloria", Presentacion: "Bloque 1 kg", Cantidad: 1, Precio: 108, Proveedor: "Lácteos del Bajío"},
	}},
	{Slug: "huevo", Nombre: "Huevo", Unidad: "pieza", MermaPct: 12, MermaOrigen: "manual", ActivoIdx: 0, DiasPrecio: 9, Productos: []Producto{
		{Marca: "San Juan", Presentacion: "Caja 30 pzas", Cantidad: 30, Precio: 49.50, Proveedor: "Mercado de Abastos"},
	}},
	{Slug: "azucar", Nombre: "Azúcar estándar", Unidad: "kg", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 20, Productos: []Producto{
		{Marca: "Zulka", Presentacion: "Bolsa 2 kg", Cantidad: 2, Precio: 55, Proveedor: "Súper del barrio"},
	}},
	{Slug: "chocolate", Nombre: "Chocolate semiamargo", Unidad: "kg", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 122, Productos: []Producto{
		{Marca: "Turin repostero", Presentacion: "Bolsa 2.5 kg", Cantidad: 2.5, Precio: 961.25, Proveedor: "Distribuidora dulcera"},
	}},
	{Slug: "pina", Nombre: "Piña", Unidad: "kg", MermaPct: 25, MermaOrigen: "medido", ActivoIdx: 0, DiasPrecio: 5, Productos: []Producto{
		{Marca: "Piña miel", Presentacion: "Por kilo", Cantidad: 1, Precio: 27.50, Proveedor: "Mercado de Abastos"},
	}},
	{Slug: "leche", Nombre: "Leche entera", Unidad: "L", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 8, Productos: []Producto{
		{Marca: "Lala", Presentacion: "Garrafa 4 L", Cantidad: 4, Precio: 92, Proveedor: "Súper del barrio"},
	}},
	{Slug: "crema", Nombre: "Crema para batir", Unidad: "L", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 30, Productos: []Producto{
		{Marca: "Lyncott", Presentacion: "Botella 1 L", Cantidad: 1, Precio: 70, Proveedor: "Súper del barrio"},
	}},
	{Slug: "fresa", Nombre: "Fresa", Unidad: "kg", MermaPct: 15, MermaOrigen: "manual", ActivoIdx: 0, DiasPrecio: 3, Productos: []Producto{
		{Marca: "Fresa fresca", Presentacion: "Caja 2 kg", Cantidad: 2, Precio: 68, Proveedor: "Mercado de Abastos"},
	}},
	{Slug: "frambuesa", Nombre: "Frambuesa", Unidad: "kg", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 15, Productos: []Producto{
		{Marca: "Congelada", Presentacion: "Bolsa 1 kg", Cantidad: 1, Precio: 161.40, Proveedor: "Distribuidora dulcera"},
	}},
	{Slug: "limon", Nombre: "Limón", Unidad: "kg", MermaPct: 30, MermaOrigen: "medido", ActivoIdx: 0, DiasPrecio: 6, Productos: []Producto{
		{Marca: "Limón sin semilla", Presentacion: "Por kilo", Cantidad: 1, Precio: 24.50, Proveedor: "Mercado de Abastos"},
	}},
	{Slug: "almendra", Nombre: "Almendra en polvo", Unidad: "kg", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 70, Productos: []Producto{
		{Marca: "Importada", Presentacion: "Bolsa 1 kg", Cantidad: 1, Precio: 318, Proveedor: "Distribuidora dulcera"},
	}},
	{Slug: "vainilla", Nombre: "Vainilla natural", Unidad: "L", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 40, Productos: []Producto{
		{Marca: "Villa Rica", Presentacion: "Frasco 250 ml", Cantidad: 0.25, Precio: 152.50, Proveedor: "Distribuidora dulcera"},
	}},
	{Slug: "canela", Nombre: "Canela molida", Unidad: "kg", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 33, Productos: []Producto{
		{Marca: "McCormick", Presentacion: "Bolsa 250 g", Cantidad: 0.25, Precio: 62.50, Proveedor: "Súper del barrio"},
	}},
	{Slug: "levadura", Nombre: "Levadura seca", Unidad: "kg", MermaPct: 0, MermaOrigen: "referencia", ActivoIdx: 0, DiasPrecio: 25, Productos: []Producto{
		{Marca: "Nevada", Presentacion: "Paquete 500 g", Cantidad: 0.5, Precio: 60, Proveedor: "Súper del barrio"},
	}},
}

// Recetas — subrecetas primero para que el seeder resuelva referencias.
var Recetas = []Receta{
	{Slug: "mousseline", Nombre: "Mousseline almendra", Categoria: "Complementos", Porciones: 1,
		Etiqueta: "Rinde 950 g", EtiquetaSingular: "kilo", RendimientoKg: 0.95, PrecioVenta: nil, EsSubreceta: true,
		Alergenos: []string{"Huevo", "Lácteos", "Frutos secos"},
		Lineas: []Linea{
			{IngredienteSlug: "almendra", Cantidad: 0.2},
			{IngredienteSlug: "mantequilla", Cantidad: 0.25},
			{IngredienteSlug: "azucar", Cantidad: 0.12},
			{IngredienteSlug: "huevo", Cantidad: 4},
			{IngredienteSlug: "leche", Cantidad: 0.2},
		},
		Pasos: []string{
			"Prepara una crema pastelera con la leche, el huevo y el azúcar; enfría a temperatura ambiente.",
			"Bate la mantequilla pomada e incorpora la crema en tres tiempos junto con la almendra en polvo.",
		}},
	{Slug: "buttercream", Nombre: "Butter cream de chocolate", Categoria: "Complementos", Porciones: 1,
		Etiqueta: "Rinde 3 kg", EtiquetaSingular: "kilo", RendimientoKg: 3, PrecioVenta: nil, EsSubreceta: true,
		Alergenos: []string{"Huevo", "Lácteos"},
		Lineas: []Linea{
			{IngredienteSlug: "chocolate", Cantidad: 0.40005},
			{IngredienteSlug: "mantequilla", Cantidad: 0.2},
			{IngredienteSlug: "huevo", Cantidad: 4},
		},
		Pasos: []string{
			"Funde el chocolate a baño maría y deja entibiar a 30 °C.",
			"Bate la mantequilla con el huevo pasteurizado e incorpora el chocolate en hilo.",
		}},
	{Slug: "choux", Nombre: "Choux", Categoria: "Pastas", Porciones: 9,
		Etiqueta: "piezas", EtiquetaSingular: "pieza", RendimientoKg: 0.9, PrecioVenta: precio(250), EsSubreceta: false,
		Alergenos: []string{"Gluten", "Huevo", "Lácteos", "Frutos secos"},
		Lineas: []Linea{
			{RecetaSlug: "mousseline", Cantidad: 0.95},
			{IngredienteSlug: "harina", Cantidad: 0.25},
			{IngredienteSlug: "mantequilla", Cantidad: 0.1},
			{IngredienteSlug: "huevo", Cantidad: 4},
			{IngredienteSlug: "leche", Cantidad: 0.25},
			{IngredienteSlug: "azucar", Cantidad: 0.0315},
		},
		Pasos: []string{
			"Hierve la leche con la mantequilla y una pizca de sal. Fuera del fuego, agrega la harina de golpe y mezcla hasta despegar del cazo.",
			"Deja entibiar y agrega los huevos uno por uno hasta lograr una pasta lisa que caiga en punta.",
			"Manguea 9 piezas y hornea a 190 °C por 30 minutos sin abrir el horno.",
			"Enfría por completo y rellena con la mousseline de almendra.",
		}},
	{Slug: "pastel", Nombre: "Pastel fresas con crema", Categoria: "Pasteles", Porciones: 8,
		Etiqueta: "rebanadas", EtiquetaSingular: "rebanada", RendimientoKg: 1.4, PrecioVenta: precio(450), EsSubreceta: false,
		Alergenos: []string{"Gluten", "Huevo", "Lácteos"},
		Lineas: []Linea{
			{IngredienteSlug: "harina", Cantidad: 0.25},
			{IngredienteSlug: "huevo", Cantidad: 6},
			{IngredienteSlug: "azucar", Cantidad: 0.2},
			{IngredienteSlug: "fresa", Cantidad: 0.3},
			{IngredienteSlug: "crema", Cantidad: 0.2004},
			{IngredienteSlug: "vainilla", Cantidad: 0.005},
		},
		Pasos: []string{
			"Bate los huevos con el azúcar hasta triplicar el volumen y envuelve la harina cernida.",
			"Hornea el bizcocho a 175 °C por 25 minutos y deja enfriar sobre rejilla.",
			"Monta la crema con la vainilla, rellena con fresas en mitades y decora la superficie.",
		}},
	{Slug: "muffin", Nombre: "Muffin limón y frambuesa", Categoria: "Muffins", Porciones: 11,
		Etiqueta: "piezas", EtiquetaSingular: "pieza", RendimientoKg: 1.1, PrecioVenta: precio(190), EsSubreceta: false,
		Alergenos: []string{"Gluten", "Huevo", "Lácteos"},
		Lineas: []Linea{
			{IngredienteSlug: "harina", Cantidad: 0.35},
			{IngredienteSlug: "mantequilla", Cantidad: 0.18},
			{IngredienteSlug: "huevo", Cantidad: 3},
			{IngredienteSlug: "azucar", Cantidad: 0.25},
			{IngredienteSlug: "limon", Cantidad: 0.1},
			{IngredienteSlug: "frambuesa", Cantidad: 0.1501},
		},
		Pasos: []string{
			"Acrema la mantequilla con el azúcar y la ralladura de limón; agrega los huevos uno a uno.",
			"Incorpora la harina alternando con el jugo de limón, sin sobrebatir.",
			"Reparte en 11 capacillos, hunde las frambuesas y hornea a 180 °C por 22 minutos.",
		}},
	{Slug: "roles", Nombre: "Roles de canela", Categoria: "Panes", Porciones: 12,
		Etiqueta: "piezas", EtiquetaSingular: "pieza", RendimientoKg: 1.5, PrecioVenta: precio(145), EsSubreceta: false,
		Alergenos: []string{"Gluten", "Huevo", "Lácteos"},
		Lineas: []Linea{
			{IngredienteSlug: "harina", Cantidad: 0.5},
			{IngredienteSlug: "mantequilla", Cantidad: 0.15},
			{IngredienteSlug: "azucar", Cantidad: 0.18},
			{IngredienteSlug: "canela", Cantidad: 0.01},
			{IngredienteSlug: "huevo", Cantidad: 2},
			{IngredienteSlug: "leche", Cantidad: 0.2},
		},
		Pasos: []string{
			"Amasa todos los ingredientes de la masa hasta que esté lisa y elástica; deja doblar su tamaño.",
			"Extiende, unta la mantequilla con azúcar y canela, enrolla y corta 12 piezas.",
			"Fermenta 45 minutos y hornea a 180 °C por 18 minutos.",
		}},
	{Slug: "croissant", Nombre: "Croissant clásico", Categoria: "Panes", Porciones: 8,
		Etiqueta: "piezas", EtiquetaSingular: "pieza", RendimientoKg: 1.0, PrecioVenta: precio(190), EsSubreceta: false,
		Alergenos: []string{"Gluten", "Lácteos"},
		Lineas: []Linea{
			{IngredienteSlug: "harina", Cantidad: 0.5},
			{IngredienteSlug: "mantequilla", Cantidad: 0.28},
			{IngredienteSlug: "leche", Cantidad: 0.14},
			{IngredienteSlug: "azucar", Cantidad: 0.06},
			{IngredienteSlug: "levadura", Cantidad: 0.01},
		},
		Pasos: []string{
			"Prepara la masa y refrigera 12 horas antes de laminar.",
			"Lamina con la mantequilla en tres dobleces sencillos, reposando entre cada uno.",
			"Forma los croissants, fermenta 2 horas y hornea a 200 °C por 16 minutos.",
		}},
	{Slug: "cupcake", Nombre: "Cupcake de vainilla", Categoria: "Cupcakes", Porciones: 12,
		Etiqueta: "piezas", EtiquetaSingular: "pieza", RendimientoKg: 1.2, PrecioVenta: precio(160), EsSubreceta: false,
		Alergenos: []string{"Gluten", "Huevo", "Lácteos"},
		Lineas: []Linea{
			{IngredienteSlug: "harina", Cantidad: 0.3},
			{IngredienteSlug: "mantequilla", Cantidad: 0.15},
			{IngredienteSlug: "azucar", Cantidad: 0.2},
			{IngredienteSlug: "huevo", Cantidad: 3},
			{IngredienteSlug: "vainilla", Cantidad: 0.008},
			{IngredienteSlug: "leche", Cantidad: 0.15},
		},
		Pasos: []string{
			"Acrema la mantequilla con el azúcar, agrega los huevos y la vainilla.",
			"Incorpora la harina alternando con la leche y reparte en 12 capacillos.",
			"Hornea a 175 °C por 20 minutos.",
		}},
}

// BuildModels materializes the dataset as in-memory model maps keyed by slug
// (no database involved) — used by the costing-engine tests.
func BuildModels() (map[string]*models.Ingrediente, map[string]*models.Receta) {
	ingredientes := map[string]*models.Ingrediente{}
	for _, seed := range Ingredientes {
		ing := &models.Ingrediente{
			ID:          seed.Slug,
			Nombre:      seed.Nombre,
			UnidadBase:  seed.Unidad,
			MermaPct:    seed.MermaPct,
			MermaOrigen: seed.MermaOrigen,
		}
		for i, p := range seed.Productos {
			ing.Productos = append(ing.Productos, models.ProductoCompra{
				ID: seed.Slug + ":" + p.Marca, IngredienteID: seed.Slug,
				Marca: p.Marca, Presentacion: p.Presentacion,
				Cantidad: p.Cantidad, Precio: p.Precio, Proveedor: p.Proveedor,
				Activo: i == seed.ActivoIdx, Orden: i,
			})
		}
		ingredientes[seed.Slug] = ing
	}

	recetas := map[string]*models.Receta{}
	for _, seed := range Recetas {
		rec := &models.Receta{
			ID: seed.Slug, Nombre: seed.Nombre, Categoria: seed.Categoria,
			Porciones: seed.Porciones, Etiqueta: seed.Etiqueta, EtiquetaSingular: seed.EtiquetaSingular,
			RendimientoKg: seed.RendimientoKg, PrecioVenta: seed.PrecioVenta,
			IvaPct: 16, EsSubreceta: seed.EsSubreceta,
			Alergenos: seed.Alergenos, Pasos: seed.Pasos,
		}
		for i, l := range seed.Lineas {
			linea := models.RecetaLinea{RecetaID: seed.Slug, Orden: i, Cantidad: l.Cantidad}
			if l.IngredienteSlug != "" {
				id := l.IngredienteSlug
				linea.IngredienteID = &id
			} else {
				id := l.RecetaSlug
				linea.SubRecetaID = &id
			}
			rec.Lineas = append(rec.Lineas, linea)
		}
		recetas[seed.Slug] = rec
	}

	return ingredientes, recetas
}
