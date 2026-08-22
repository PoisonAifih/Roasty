package services

import (
	"encoding/json"
	"fmt"
	"strings"
)

// BeanSupplier is one contact point for sourcing a specific bean.
type BeanSupplier struct {
	Name     string `json:"name"`
	Type     string `json:"type"`     // koperasi | pedagang | online
	Contact  string `json:"contact"`  // WhatsApp / phone
	Location string `json:"location"`
	MinOrder string `json:"min_order"`
	Note     string `json:"note"`
}

// BeanCatalogEntry is a rich bean profile stored in Qdrant for semantic search.
// SQL holds the numbers; this holds the context, profile, and sourcing map.
type BeanCatalogEntry struct {
	ID            string
	Origin        string
	Variety       string
	Grade         string // specialty | premium | commercial
	PriceRange    string // Rp X–Y/kg, buyer-level price
	FlavorNotes   []string
	BodyAcidity   string
	Processing    string
	Altitude      string
	HarvestSeason string
	UseCases      []string
	Suppliers     []BeanSupplier
}

// document builds the narrative text that gets embedded into Qdrant.
// Rich prose makes semantic similarity work well: queries like "fruity low-acid
// arabika for pour over" land on the right entries even with no keyword overlap.
func (e BeanCatalogEntry) document() string {
	return fmt.Sprintf(
		`%s %s dari %s.
Grade: %s. Harga: %s.
Profil rasa: %s. %s.
Proses: %s. Ketinggian: %s. Panen: %s.
Cocok untuk: %s.`,
		e.Variety, e.ID, e.Origin,
		e.Grade, e.PriceRange,
		strings.Join(e.FlavorNotes, ", "), e.BodyAcidity,
		e.Processing, e.Altitude, e.HarvestSeason,
		strings.Join(e.UseCases, "; "),
	)
}

// resultJSON serialises the entry as the string the agent reads back from the
// find_similar_beans tool — structured so the model can cite suppliers directly.
func (e BeanCatalogEntry) resultJSON() string {
	type out struct {
		Origin        string         `json:"origin"`
		Variety       string         `json:"variety"`
		Grade         string         `json:"grade"`
		PriceRange    string         `json:"price_range"`
		FlavorNotes   []string       `json:"flavor_notes"`
		BodyAcidity   string         `json:"body_acidity"`
		Processing    string         `json:"processing"`
		HarvestSeason string         `json:"harvest_season"`
		UseCases      []string       `json:"use_cases"`
		Suppliers     []BeanSupplier `json:"suppliers"`
	}
	b, _ := json.Marshal(out{
		Origin:        e.Origin,
		Variety:       e.Variety,
		Grade:         e.Grade,
		PriceRange:    e.PriceRange,
		FlavorNotes:   e.FlavorNotes,
		BodyAcidity:   e.BodyAcidity,
		Processing:    e.Processing,
		HarvestSeason: e.HarvestSeason,
		UseCases:      e.UseCases,
		Suppliers:     e.Suppliers,
	})
	return string(b)
}

// BeanCatalog is the full sourcing catalog seeded into Qdrant.
// It intentionally includes beans not currently in stock — the point is to
// surface procurement options the roastery hasn't considered yet.
var BeanCatalog = []BeanCatalogEntry{

	// ── Arabika Specialty ────────────────────────────────────────────────────

	{
		ID:            "gayo-washed",
		Origin:        "Gayo, Aceh",
		Variety:       "Arabica",
		Grade:         "specialty",
		PriceRange:    "Rp 85.000–100.000/kg",
		FlavorNotes:   []string{"herbal", "dark chocolate", "earthy", "caramel", "aftertaste panjang"},
		BodyAcidity:   "body tebal, asam medium-rendah",
		Processing:    "giling basah (wet-hulled)",
		Altitude:      "1.200–1.500 mdpl",
		HarvestSeason: "April–Juni (utama), Oktober–Desember (minor)",
		UseCases:      []string{"single origin espresso", "pour over / V60", "blend arabika premium"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Gayo Organik Maju", Type: "koperasi", Contact: "0812-6001-1101", Location: "Bener Meriah, Aceh", MinOrder: "50 kg", Note: "bersertifikat organik, tersedia grade 1 & 2"},
			{Name: "CV. Aceh Prima Kopi", Type: "pedagang", Contact: "0821-7701-2201", Location: "Banda Aceh", MinOrder: "10 kg", Note: "ready stock, bisa kirim ke Jawa dalam 3 hari"},
		},
	},
	{
		ID:            "gayo-natural",
		Origin:        "Gayo, Aceh — Natural Process",
		Variety:       "Arabica",
		Grade:         "specialty",
		PriceRange:    "Rp 100.000–130.000/kg",
		FlavorNotes:   []string{"fruity intens", "blueberry", "wine-like", "dark chocolate", "manis"},
		BodyAcidity:   "body sangat tebal, asam cerah fruity",
		Processing:    "natural (buah dikeringkan utuh)",
		Altitude:      "1.200–1.500 mdpl",
		HarvestSeason: "April–Juni",
		UseCases:      []string{"single origin pour over premium", "cold brew fruity", "pasar specialty enthusiast"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Gayo Organik Maju", Type: "koperasi", Contact: "0812-6001-1101", Location: "Bener Meriah, Aceh", MinOrder: "25 kg", Note: "lot natural terbatas, pesan 3 bulan sebelum panen"},
			{Name: "Rumah Kopi Gayo", Type: "pedagang", Contact: "0813-6101-2202", Location: "Takengon, Aceh Tengah", MinOrder: "10 kg", Note: "spesialis natural & honey process Gayo"},
		},
	},
	{
		ID:            "toraja-sulsel",
		Origin:        "Toraja, Sulawesi Selatan",
		Variety:       "Arabica",
		Grade:         "specialty",
		PriceRange:    "Rp 90.000–115.000/kg",
		FlavorNotes:   []string{"earthy", "wine-like", "dark chocolate", "herbal", "spicy"},
		BodyAcidity:   "body sangat tebal, asam rendah",
		Processing:    "giling basah (wet-hulled)",
		Altitude:      "1.000–1.500 mdpl",
		HarvestSeason: "Juni–September",
		UseCases:      []string{"single origin espresso", "blend premium", "pasar Jepang & Korea"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Petani Toraja Rindang", Type: "koperasi", Contact: "0812-4201-3301", Location: "Rantepao, Toraja Utara", MinOrder: "50 kg", Note: "grade 1 & 2, suplai ke eksportir Jepang"},
			{Name: "UD. Nusantara Kopi Sulawesi", Type: "pedagang", Contact: "0821-5501-4401", Location: "Makassar", MinOrder: "20 kg", Note: "middleman terpercaya, harga nego untuk volume"},
		},
	},
	{
		ID:            "toraja-honey",
		Origin:        "Toraja, Sulawesi Selatan — Honey Process",
		Variety:       "Arabica",
		Grade:         "specialty",
		PriceRange:    "Rp 105.000–130.000/kg",
		FlavorNotes:   []string{"fruity", "manis", "caramel", "dark cherry", "floral ringan"},
		BodyAcidity:   "body tebal, asam medium cerah",
		Processing:    "honey process",
		Altitude:      "1.000–1.500 mdpl",
		HarvestSeason: "Juni–September",
		UseCases:      []string{"single origin pour over", "pasar specialty premium", "espresso fruity"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Petani Toraja Rindang", Type: "koperasi", Contact: "0812-4201-3301", Location: "Rantepao, Toraja Utara", MinOrder: "25 kg", Note: "lot honey terbatas, tersedia Juni–Agustus"},
		},
	},
	{
		ID:            "flores-bajawa",
		Origin:        "Flores Bajawa, NTT",
		Variety:       "Arabica",
		Grade:         "specialty",
		PriceRange:    "Rp 88.000–110.000/kg",
		FlavorNotes:   []string{"floral kuat", "fruity", "caramel", "citrus", "dark chocolate"},
		BodyAcidity:   "body medium, asam cerah",
		Processing:    "semi-washed / fully-washed",
		Altitude:      "1.000–1.800 mdpl",
		HarvestSeason: "Juni–Agustus",
		UseCases:      []string{"single origin pour over", "cold brew floral", "pasar specialty enthusiast"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Petani Flores Bajawa", Type: "koperasi", Contact: "0813-8101-5501", Location: "Bajawa, Ngada, NTT", MinOrder: "30 kg", Note: "direct trade, bisa visit kebun"},
			{Name: "UD. Flores Kopi Asli", Type: "pedagang", Contact: "0822-9101-6601", Location: "Ende, Flores", MinOrder: "10 kg", Note: "stock tersedia, kirim via Surabaya"},
		},
	},
	{
		ID:            "flores-manggarai",
		Origin:        "Flores Manggarai, NTT",
		Variety:       "Arabica",
		Grade:         "premium",
		PriceRange:    "Rp 80.000–100.000/kg",
		FlavorNotes:   []string{"earthy", "herbal", "coklat hitam", "nutty", "sedikit fruity"},
		BodyAcidity:   "body tebal, asam rendah-medium",
		Processing:    "giling basah / semi-washed",
		Altitude:      "800–1.400 mdpl",
		HarvestSeason: "Juli–September",
		UseCases:      []string{"blend arabika", "espresso medium roast", "pasar domestik premium"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Manggarai Arabica", Type: "koperasi", Contact: "0813-7201-7701", Location: "Ruteng, Manggarai, NTT", MinOrder: "50 kg", Note: "alternatif lebih murah dari Bajawa, profil mirip"},
			{Name: "UD. Flores Kopi Asli", Type: "pedagang", Contact: "0822-9101-6601", Location: "Ende, Flores", MinOrder: "10 kg", Note: "bisa campur Bajawa & Manggarai dalam satu order"},
		},
	},
	{
		ID:            "java-preanger",
		Origin:        "Java Preanger, Jawa Barat",
		Variety:       "Arabica",
		Grade:         "specialty",
		PriceRange:    "Rp 92.000–120.000/kg",
		FlavorNotes:   []string{"bersih (clean)", "nutty", "coklat susu", "citrus ringan", "manis"},
		BodyAcidity:   "body medium, asam cerah (brightness tinggi)",
		Processing:    "fully-washed",
		Altitude:      "1.000–1.800 mdpl",
		HarvestSeason: "Agustus–Oktober",
		UseCases:      []string{"single origin pour over / V60", "filter coffee", "pasar Eropa & Australia"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Java Preanger Pangalengan", Type: "koperasi", Contact: "0812-2301-8801", Location: "Pengalengan, Bandung Selatan", MinOrder: "50 kg", Note: "Grade 1 bersertifikat, tersedia lot micro"},
			{Name: "UD. Jabar Kopi Prima", Type: "pedagang", Contact: "0821-3301-9901", Location: "Bandung", MinOrder: "10 kg", Note: "siap kirim Bandung-Jakarta-Surabaya"},
		},
	},
	{
		ID:            "wamena-papua",
		Origin:        "Wamena, Papua",
		Variety:       "Arabica",
		Grade:         "specialty",
		PriceRange:    "Rp 105.000–155.000/kg",
		FlavorNotes:   []string{"floral", "fruity", "sangat bersih (clean cup)", "manis", "aftertaste panjang"},
		BodyAcidity:   "body medium-light, asam cerah dan bersih",
		Processing:    "fully-washed (organik)",
		Altitude:      "1.500–2.000 mdpl",
		HarvestSeason: "Oktober–Desember",
		UseCases:      []string{"single origin ultra-premium", "cupping showcase", "pasar Jepang & specialty high-end"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Organik Wamena", Type: "koperasi", Contact: "0812-9801-1001", Location: "Wamena, Jayawijaya, Papua", MinOrder: "25 kg", Note: "supply terbatas, pesan jauh-jauh hari; bersertifikat organik"},
			{Name: "PT. Papua Organic Coffee", Type: "pedagang", Contact: "0821-8801-2002", Location: "Jayapura, Papua", MinOrder: "10 kg", Note: "aggregator Papua, bisa sourcing lot-lot kecil dari petani"},
		},
	},
	{
		ID:            "kintamani-bali",
		Origin:        "Kintamani, Bali",
		Variety:       "Arabica",
		Grade:         "premium",
		PriceRange:    "Rp 75.000–95.000/kg",
		FlavorNotes:   []string{"citrusy", "lemon", "jeruk keprok", "fruity", "clean"},
		BodyAcidity:   "body medium, asam cerah dan segar",
		Processing:    "fully-washed",
		Altitude:      "900–1.500 mdpl",
		HarvestSeason: "Juni–September",
		UseCases:      []string{"cold brew citrusy", "pour over light roast", "pasar Bali & Australia"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Subak Abian Kintamani", Type: "koperasi", Contact: "0812-1801-3003", Location: "Kintamani, Bangli, Bali", MinOrder: "30 kg", Note: "IG terdaftar, tersedia versi organik"},
			{Name: "CV. Bali Kopi Premium", Type: "pedagang", Contact: "0821-0201-4004", Location: "Denpasar, Bali", MinOrder: "5 kg", Note: "min. order paling kecil, cocok untuk trial"},
		},
	},

	// ── Arabika Premium / Menengah ───────────────────────────────────────────

	{
		ID:            "mandailing-sumut",
		Origin:        "Mandailing, Sumatera Utara",
		Variety:       "Arabica",
		Grade:         "premium",
		PriceRange:    "Rp 72.000–90.000/kg",
		FlavorNotes:   []string{"earthy", "herbal", "dark chocolate", "body tebal", "asam rendah"},
		BodyAcidity:   "body sangat tebal, asam rendah (karakter Sumatera khas)",
		Processing:    "giling basah (wet-hulled)",
		Altitude:      "800–1.400 mdpl",
		HarvestSeason: "April–Juni",
		UseCases:      []string{"espresso blend arabika Sumatera", "single origin untuk pasar AS", "blend body-forward"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Mandailing Arabica Sejahtera", Type: "koperasi", Contact: "0812-5501-5005", Location: "Panyabungan, Mandailing Natal", MinOrder: "50 kg", Note: "dikenal di pasar ekspor AS sebagai Sumatran Mandheling"},
			{Name: "CV. Sumut Kopi Jaya", Type: "pedagang", Contact: "0821-6601-6006", Location: "Medan, Sumatera Utara", MinOrder: "10 kg", Note: "aggregator Sumatera Utara, bisa campur Mandailing + Sidikalang"},
		},
	},
	{
		ID:            "sidikalang-sumut",
		Origin:        "Sidikalang, Sumatera Utara",
		Variety:       "Arabica",
		Grade:         "premium",
		PriceRange:    "Rp 68.000–85.000/kg",
		FlavorNotes:   []string{"earthy", "herbal", "dark fruit", "body tebal", "sedikit spicy"},
		BodyAcidity:   "body tebal, asam medium-rendah",
		Processing:    "giling basah (wet-hulled)",
		Altitude:      "1.000–1.400 mdpl",
		HarvestSeason: "April–Juni",
		UseCases:      []string{"blend arabika Sumatera terjangkau", "espresso house blend", "alternatif lebih murah dari Gayo"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Sidikalang Arabica", Type: "koperasi", Contact: "0812-7701-7007", Location: "Sidikalang, Dairi, Sumut", MinOrder: "30 kg", Note: "harga lebih murah dari Gayo, profil mirip"},
			{Name: "CV. Sumut Kopi Jaya", Type: "pedagang", Contact: "0821-6601-6006", Location: "Medan, Sumatera Utara", MinOrder: "10 kg", Note: "siap kirim seluruh Indonesia"},
		},
	},
	{
		ID:            "lintong-sumut",
		Origin:        "Lintong, Sumatera Utara",
		Variety:       "Arabica",
		Grade:         "premium",
		PriceRange:    "Rp 70.000–88.000/kg",
		FlavorNotes:   []string{"earthy", "herbal", "nutty", "coklat", "clean cup lebih baik dari Mandailing"},
		BodyAcidity:   "body tebal, asam medium",
		Processing:    "giling basah (wet-hulled)",
		Altitude:      "1.100–1.600 mdpl",
		HarvestSeason: "Maret–Mei",
		UseCases:      []string{"blend arabika Sumatera", "single origin untuk pasar yang suka karakter Sumatera tapi lebih bersih"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Lintong Jaya Arabica", Type: "koperasi", Contact: "0812-3301-8008", Location: "Lintongnihuta, Humbang Hasundutan, Sumut", MinOrder: "50 kg", Note: "sering diblending dengan Mandailing untuk kompleksitas"},
			{Name: "CV. Sumut Kopi Jaya", Type: "pedagang", Contact: "0821-6601-6006", Location: "Medan", MinOrder: "10 kg", Note: "stok bersama Sidikalang & Mandailing"},
		},
	},
	{
		ID:            "enrekang-sulsel",
		Origin:        "Enrekang, Sulawesi Selatan",
		Variety:       "Arabica",
		Grade:         "premium",
		PriceRange:    "Rp 80.000–100.000/kg",
		FlavorNotes:   []string{"earthy", "dark chocolate", "caramel", "sedikit fruity", "herbal"},
		BodyAcidity:   "body tebal, asam medium-rendah",
		Processing:    "giling basah / semi-washed",
		Altitude:      "1.000–1.400 mdpl",
		HarvestSeason: "Mei–Agustus",
		UseCases:      []string{"alternatif Toraja lebih murah", "blend espresso premium", "single origin Sulawesi"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Petani Enrekang Arabica", Type: "koperasi", Contact: "0812-0401-9009", Location: "Enrekang, Sulawesi Selatan", MinOrder: "30 kg", Note: "profil mirip Toraja, harga 10–15% lebih murah"},
			{Name: "UD. Nusantara Kopi Sulawesi", Type: "pedagang", Contact: "0821-5501-4401", Location: "Makassar", MinOrder: "10 kg", Note: "campur Toraja + Enrekang tersedia"},
		},
	},

	// ── Robusta ──────────────────────────────────────────────────────────────

	{
		ID:            "lampung-robusta",
		Origin:        "Lampung, Sumatera",
		Variety:       "Robusta",
		Grade:         "commercial",
		PriceRange:    "Rp 38.000–52.000/kg",
		FlavorNotes:   []string{"pahit kuat", "earthy", "cocoa", "smokiness ringan"},
		BodyAcidity:   "body sangat tebal, asam sangat rendah, kafein tinggi",
		Processing:    "semi-washed / dry",
		Altitude:      "200–600 mdpl",
		HarvestSeason: "Mei–Juli",
		UseCases:      []string{"base blend espresso volume tinggi", "kopi instan", "blend untuk body dan kafein"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Robusta Lampung Barat", Type: "koperasi", Contact: "0812-1101-0010", Location: "Liwa, Lampung Barat", MinOrder: "100 kg", Note: "volume besar, harga bisa nego; grade 1 & bulk"},
			{Name: "UD. Lampung Kopi Express", Type: "pedagang", Contact: "0821-2201-1011", Location: "Bandar Lampung", MinOrder: "25 kg", Note: "paling cepat untuk restock volume robusta"},
		},
	},
	{
		ID:            "temanggung-robusta",
		Origin:        "Temanggung, Jawa Tengah",
		Variety:       "Robusta",
		Grade:         "commercial",
		PriceRange:    "Rp 45.000–62.000/kg",
		FlavorNotes:   []string{"pahit premium", "tembakau", "coklat dark", "earthy khas Jawa"},
		BodyAcidity:   "body tebal, asam rendah, aftertaste smoky",
		Processing:    "semi-washed",
		Altitude:      "500–900 mdpl",
		HarvestSeason: "Agustus–Oktober",
		UseCases:      []string{"blend espresso premium Jawa", "kopi tubruk robusta premium", "blend dengan arabika untuk balance"},
		Suppliers: []BeanSupplier{
			{Name: "Koperasi Robusta Temanggung", Type: "koperasi", Contact: "0812-3401-2012", Location: "Temanggung, Jawa Tengah", MinOrder: "50 kg", Note: "robusta Jawa paling premium, lebih mahal dari Lampung tapi profil unik"},
			{Name: "UD. Jateng Kopi Prima", Type: "pedagang", Contact: "0821-4501-3013", Location: "Semarang, Jawa Tengah", MinOrder: "10 kg", Note: "kirim ke Jawa 1–2 hari"},
		},
	},
}
