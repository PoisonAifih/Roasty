# Dataset References

Sumber data untuk Roasty, hasil riset 11 Agustus 2026.

**Aturan pembagian** — ini menentukan sebuah dataset masuk ke mana:

| Jenis | Contoh | Tujuan |
|---|---|---|
| Angka / terstruktur | harga, produksi, skor cupping | **Tabel Postgres → tool agent** |
| Teks / tidak terstruktur | standar mutu, panduan proses | **Qdrant → RAG** |

Vector search buruk untuk angka. Kalau ditanya "biji mana marginnya paling tinggi", embedding mengembalikan yang *mirip secara semantik*, bukan yang benar. SQL menjawab itu dengan pasti. Jangan tukar tempat.

---

## A. Data terstruktur → Postgres + tool

### A1. CQI Coffee Quality Database ⭐ prioritas utama

**Simplenya:** ganti angka karangan di seed dengan skor cupping asli dari penilai bersertifikat.

- **Link:** https://github.com/jldbc/coffee-quality-database
- **Isi:** 1.340 review — 1.312 arabika + 28 robusta
- **Lisensi:** MIT — bebas dipakai
- **Format:** CSV, tersedia versi mentah dan bersih (pakai yang bersih)

**Kolom:**
- *Kualitas:* Aroma, Flavor, Aftertaste, Acidity, Body, Balance, Uniformity, Cup Cleanliness, Sweetness, Moisture, Defects
- *Biji:* Processing Method, Color, Species
- *Kebun:* Owner, Country of Origin, Farm Name, Altitude, Region

**Kenapa paling cocok:** hampir semua kolom tabel `beans` ada padanannya di sini — `quality_score`, `humidity`, `variety`, `origin`. Bedanya datanya nyata. Bonus tiga dimensi yang belum kita punya: **metode proses**, **ketinggian**, **nilai cacat**.

**Catatan:** di-scrape Januari 2018. Ada versi 2023 (1.509 record, protokol SCA) beredar di ekosistem yang sama. Untuk data terbaru langsung dari CQI butuh kredensial login.

---

### A2. FAOSTAT — produksi global

**Simplenya:** konteks makro, buat jawab "produksi Indonesia naik atau turun dibanding negara lain".

- **Link:** https://www.fao.org/faostat/en/#data/QC
- **Isi:** produksi kopi per negara, **1961–2022**
- **Lisensi:** gratis
- **Format:** bulk CSV dalam zip, kode dataset `QCL`

**Prioritas:** rendah. Menarik tapi tidak langsung mengubah keputusan beli seorang roastery.

---

### A3. World Bank Pink Sheet — harga komoditas

**Simplenya:** ini yang bikin agent bisa jawab **"Rp 85.000/kg itu mahal atau murah?"** — sekarang Roasty tidak punya pembanding apa pun.

- **Link:** https://thedocs.worldbank.org/en/doc/74e8be41ceb20fa0da750cda2f6b9e4e-0050012026/related/CMO-Pink-Sheet-June-2026.pdf
- **Isi:** harga bulanan arabika & robusta (memakai indicator price ICO)
- **Lisensi:** gratis, **wajib mencantumkan atribusi**
- **Format:** PDF bulanan; versi Excel ada di situs Commodity Markets Outlook

**Prioritas:** tinggi setelah A1. Membuka pertanyaan yang paling sering ditanya pembeli.

---

### A4. BPS / Satu Data Pertanian — produksi Indonesia

**Simplenya:** ganti `harvest_estimate_kg` yang sekarang angka karangan dengan produksi provinsi yang nyata.

- **Link:** https://satudata.pertanian.go.id/details/publikasi/527
- **Outlook Kopi 2023:** https://satudata.pertanian.go.id/assets/docs/publikasi/Buku_Outlook_Kopi_2023_lengkap.pdf
- **Isi:** produksi & luas tanam kopi per provinsi
- **Lisensi:** publikasi pemerintah, gratis
- **Format:** PDF (perlu ekstraksi tabel)

**Angka 2022 sebagai gambaran:**

| Provinsi | Produksi | Porsi nasional |
|---|---|---|
| Sumatera Selatan | 212,4 rb ton | 26,72% |
| Lampung | 124,5 rb ton | — |
| Sumatera Utara | 87 rb ton | — |
| Aceh | 75,3 rb ton | — |

**Catatan:** dokumen ini dipakai dua kali — tabelnya ke Postgres, narasinya ke Qdrant.

---

### A5. ICO — ⚠️ sebagian besar tertutup

**Simplenya:** data paling lengkap yang ada, tapi kemungkinan besar tidak bisa kita akses tepat waktu.

- **Link:** https://ico.org/resources/historical-data-on-the-global-coffee-trade/
- **Isi:** World Coffee Statistics Database — bulanan sejak **Oktober 1963**
- **Akses:** **tertutup**, hanya anggota ICO atau pelanggan berbayar

**Jalur yang mungkin:** halaman mereka menyebut data Excel bisa diminta **gratis untuk peneliti akademik** lewat `stats@ico.org`. Sebagai mahasiswa layak dicoba — tapi **jangan dijadikan jalur utama**, balasannya tidak pasti tepat waktu. Pakai Pink Sheet (A3) sebagai gantinya; sumber angkanya sama.

---

## B. Teks → Qdrant (RAG)

### B1. SNI 01-2907-2008 ⭐ prioritas utama

**Simplenya:** standar resmi Indonesia untuk mutu biji kopi hijau. Model umum tidak hafal isinya, dan ini bikin jawaban agent terdengar seperti orang yang benar-benar paham bisnis kopi.

- **Link:** https://www.cctcid.com/wp-content/uploads/2018/08/SNI_2907-2008_Biji_Kopi-1.pdf
- **Format:** PDF publik

**Sistem grading berbasis nilai cacat** (sampel 300 gram):

| Grade | Nilai cacat |
|---|---|
| 1 | maks 11 |
| 2 | 12–25 |
| 3 | 26–44 |
| 4a | 45–60 |
| 4b | 61–80 |
| 5 | 81–150 |
| 6 | 151–225 |

**Syarat umum:** bebas serangga hidup, tidak berbau busuk/kapang, **kadar air maks 12,5%**, kadar kotoran maks 0,5%.

---

### B2. Outlook Kopi (Kementan)

**Simplenya:** narasi resmi kondisi perkopian Indonesia — tren, tantangan, proyeksi. Bahasa Indonesia, jadi cocok kalau agent nanti perlu menjawab dalam bahasa Indonesia.

- **Link:** https://satudata.pertanian.go.id/assets/docs/publikasi/Buku_Outlook_Kopi_2023_lengkap.pdf
- **Terbit:** tahunan

---

### B3. AEKI-AICE — mutu kopi ekspor

**Simplenya:** sudut pandang eksportir soal mutu. Melengkapi SNI dengan praktik lapangan.

- **Link:** https://www.aeki-aice.org/mutu-kopi/

---

### B4. CCTC — perbandingan standar grading

**Simplenya:** menjelaskan beda SNI vs SCA vs standar lain. Berguna saat agent harus membandingkan sistem penilaian.

- **Link:** https://www.cctcid.com/2018/08/29/beberapa-standard-pemeringkatan-mutu-biji-kopi-2/

---

### B5. Jurnal spesifik Indonesia

**Simplenya:** kedalaman domain — kenapa kopi Sumatra beda, efek giling basah, penanda mutu Gayo.

- Metode proses kopi Gayo: https://www.ncbi.nlm.nih.gov/pmc/articles/PMC10706735/
- Penanda mutu arabika specialty Indonesia: https://www.ncbi.nlm.nih.gov/pmc/articles/PMC10600306/

**Catatan:** ambil bagian yang relevan saja. Metodologi laboratorium tidak berguna untuk pemilik roastery.

---

## ⚠️ Isu yang ditemukan saat riset

**Kolom `humidity` kemungkinan salah satuan.**

SNI menetapkan kadar air biji kopi **maksimal 12,5%**. Tapi `schema.sql` mengisi `humidity` dengan **65–80**, dan `scoreBean` memperlakukan **65 sebagai optimal**:

```go
humidityScore := clamp(100-math.Abs(b.Humidity-65)*2, 0, 100)
```

- Kalau maksudnya **kadar air biji** → angkanya salah total; 72% berarti biji basah dan berjamur
- Kalau maksudnya **kelembapan gudang (%RH)** → angkanya wajar, tapi penamaannya menyesatkan karena berdiri sejajar dengan atribut biji lain

Belum diubah — menunggu keputusan. Menyelaraskan ke SNI (optimal ~11%, maks 12,5%) membuat skor bisa dirujuk ke standar resmi. Kolom `Moisture` di CQI (A1) memakai satuan yang benar, jadi impor A1 sekalian menyelesaikan ini.

---

## Urutan pengerjaan

1. **A1 CQI** → isi ulang tabel `beans`, sekalian benahi `humidity`
2. **B1 SNI** → dokumen pertama masuk Qdrant, sekaligus uji pipeline RAG
3. **A3 Pink Sheet** → tabel harga + tool `price_context`
4. **B2 Outlook Kopi** → tambah korpus
5. Sisanya kalau waktu masih ada

**Peringatan:** jangan habiskan waktu di ETL. Juri menilai demo, bukan pipeline.

---

## Kewajiban atribusi

Cantumkan di README dan slide:

- Coffee Quality Institute via jldbc/coffee-quality-database (MIT)
- World Bank Commodity Markets Outlook (Pink Sheet)
- FAOSTAT, FAO
- BPS / Kementerian Pertanian RI
- BSN — SNI 01-2907-2008

Kalau ada data yang dibangkitkan sintetis (misalnya riwayat penjualan), **sebut jelas di presentasi** bahwa itu simulasi.
