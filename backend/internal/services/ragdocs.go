package services

// KnowledgeChunks is the complete corpus loaded into Qdrant at seed time.
// Sources: SNI 01-2907-2008, Outlook Kopi 2023 (Kementan), AEKI-AICE,
// CCTC grading comparison, penelitian NCBI, dan pengetahuan industri.
var KnowledgeChunks = []KnowledgeChunk{

	// ── B1: SNI 01-2907-2008 ─────────────────────────────────────────────────

	{
		ID:     "sni-overview",
		Source: "SNI 01-2907-2008",
		Title:  "Sistem Grading Biji Kopi Indonesia",
		Text: `Standar Nasional Indonesia SNI 01-2907-2008 menetapkan sistem klasifikasi mutu biji kopi hijau (green bean) berdasarkan nilai cacat. Sampel uji menggunakan 300 gram biji kopi. Semakin tinggi nilai cacat, semakin rendah grade. Sistem ini berlaku untuk kopi arabika maupun robusta dan diakui oleh Badan Standardisasi Nasional (BSN). SNI menjadi acuan wajib perdagangan kopi dalam negeri dan ekspor Indonesia.`,
	},
	{
		ID:     "sni-grade-table",
		Source: "SNI 01-2907-2008",
		Title:  "Tabel Grade dan Nilai Cacat SNI",
		Text: `Tabel grade biji kopi SNI 01-2907-2008 per 300 gram sampel:
Grade 1: nilai cacat maksimal 11 — mutu tertinggi, eligible specialty
Grade 2: nilai cacat 12–25
Grade 3: nilai cacat 26–44
Grade 4a: nilai cacat 45–60
Grade 4b: nilai cacat 61–80
Grade 5: nilai cacat 81–150
Grade 6: nilai cacat 151–225 — mutu terendah yang masih dapat diperdagangkan
Biji dengan cacat di atas 225 tidak layak diperjualbelikan.`,
	},
	{
		ID:     "sni-moisture",
		Source: "SNI 01-2907-2008",
		Title:  "Kadar Air Biji Kopi Menurut SNI",
		Text: `SNI 01-2907-2008 menetapkan kadar air (moisture content) biji kopi hijau maksimal 12,5 persen. Kadar air optimal untuk penyimpanan jangka panjang adalah 10–12 persen. Biji dengan kadar air di atas 12,5% berisiko tumbuh jamur, mengalami fermentasi liar, dan degradasi rasa yang signifikan. Pengukuran dilakukan dengan moisture meter pada biji. Kelembapan gudang (relative humidity / %RH) adalah hal berbeda dan biasanya dijaga di 50–70%RH — tidak boleh dicampur dengan kadar air biji.`,
	},
	{
		ID:     "sni-general-requirements",
		Source: "SNI 01-2907-2008",
		Title:  "Syarat Umum Mutu Biji Kopi SNI",
		Text: `Syarat umum mutu biji kopi hijau menurut SNI 01-2907-2008: (1) bebas dari serangga hidup, (2) tidak berbau busuk, asam, atau kapang, (3) kadar air maksimal 12,5%, (4) kadar kotoran (benda asing) maksimal 0,5% dari berat sampel. Biji yang melanggar syarat umum ini tidak dapat diperdagangkan berapapun grade-nya, bahkan jika nilai cacat fisiknya rendah.`,
	},
	{
		ID:     "sni-defect-values",
		Source: "SNI 01-2907-2008",
		Title:  "Nilai Cacat per Jenis Cacat (Defect Reference)",
		Text: `Nilai cacat per jenis untuk menghitung grade SNI: biji hitam penuh = 1 cacat; biji hitam sebagian = 0,5; biji hitam pecah = 0,5; biji berlubang satu = 0,2; biji berlubang lebih dari satu = 0,5; biji bertutul-tutul = 0,2; biji coklat = 0,25; biji pecah = 0,2; biji muda (kulit ari) = 0,2; biji berlubang akibat serangga = 0,1; biji berkulit tanduk = 0,5; biji gelap = 0,25; ranting kecil = 5; tanah/batu kecil = 5. Total nilai cacat pada 300 gram sampel menentukan grade akhir.`,
	},

	// ── B2: Outlook Kopi 2023 (Kementan / BPS) ───────────────────────────────

	{
		ID:     "outlook-national-production",
		Source: "Outlook Kopi 2023, Kementan RI",
		Title:  "Produksi Kopi Indonesia Nasional",
		Text: `Total produksi kopi Indonesia 2022 sekitar 794 ribu ton. Indonesia adalah produsen kopi terbesar keempat di dunia setelah Brasil, Vietnam, dan Kolombia. Sekitar 75% produksi nasional adalah kopi robusta, sisanya arabika. Produktivitas rata-rata kebun rakyat masih rendah, sekitar 700–800 kg per hektar per tahun. Total luas areal kopi sekitar 1,2 juta hektar, mayoritas (96%) dikelola perkebunan rakyat.`,
	},
	{
		ID:     "outlook-province-production",
		Source: "Outlook Kopi 2023, Kementan RI",
		Title:  "Sentra Produksi Kopi per Provinsi 2022",
		Text: `Produksi kopi per provinsi Indonesia 2022: Sumatera Selatan 212,4 ribu ton (26,72% nasional — terbesar); Lampung 124,5 ribu ton; Sumatera Utara 87 ribu ton (termasuk Gayo Lues, Mandailing); Aceh 75,3 ribu ton (Gayo); Sulawesi Selatan dan Tengah (Toraja, Enrekang); NTT dan Flores (Bajawa); Jawa Barat (Preanger). Kalimantan dan Papua produksinya kecil tapi kualitasnya premium (Wamena).`,
	},
	{
		ID:     "outlook-export",
		Source: "Outlook Kopi 2023, Kementan RI",
		Title:  "Ekspor Kopi Indonesia",
		Text: `Ekspor kopi Indonesia 2022 sekitar 384 ribu ton senilai USD 1,1 miliar. Tujuan ekspor utama: Amerika Serikat, Jerman, Malaysia, Italia, dan Jepang. Robusta diekspor besar-besaran sebagai green bean untuk industri kopi instan global. Arabika specialty mulai diekspor sebagai roasted bean atau dengan value-added processing. Harga ekspor arabika specialty bisa 3–5x lipat harga robusta.`,
	},
	{
		ID:     "outlook-domestic-consumption",
		Source: "Outlook Kopi 2023, Kementan RI",
		Title:  "Konsumsi Kopi Domestik dan Tren Pasar",
		Text: `Konsumsi kopi domestik Indonesia meningkat konsisten. Pertumbuhan kafe dan kedai kopi specialty 15–20% per tahun di kota besar. Konsumen kota semakin paham origin dan proses — single origin, natural process, honey process diminati. Permintaan green bean graded meningkat, mendorong pertumbuhan roastery kecil dan menengah. Konsumsi per kapita masih rendah (sekitar 1,5 kg/tahun) dibanding Finlandia (12 kg) — menunjukkan potensi pertumbuhan besar.`,
	},

	// ── B3: AEKI-AICE (Mutu kopi ekspor) ─────────────────────────────────────

	{
		ID:     "aeki-arabika-grade",
		Source: "AEKI-AICE, Mutu Kopi Ekspor",
		Title:  "Grade Arabika untuk Ekspor",
		Text: `Menurut AEKI (Asosiasi Eksportir dan Industri Kopi Indonesia), arabika ekspor terbaik adalah Grade 1 (specialty) dengan nilai cacat SNI maksimal 11 per 300g dan cupping score ≥80. Grade 2 masih diterima pasar premium. Arabika Gayo, Mandheling (Mandailing), Toraja, dan Java Preanger dikenal luas di pasar internasional. Kopi specialty grade 1 dari Gayo bisa mencapai USD 5–8 per kg FOB, sedangkan grade komersial USD 2–3 per kg.`,
	},
	{
		ID:     "aeki-robusta-grade",
		Source: "AEKI-AICE, Mutu Kopi Ekspor",
		Title:  "Grade Robusta untuk Ekspor",
		Text: `Robusta Indonesia (terutama Lampung dan Sumatera Selatan) diekspor sebagai EK-1 hingga EK-4. Robusta Lampung terkenal rasa kuat, pahit, earthy, cocok untuk blend espresso dan kopi instan. Harga robusta jauh di bawah arabika, biasanya USD 1,5–2,5 per kg FOB atau Rp 35.000–55.000/kg di tingkat pedagang lokal. Permintaan robusta stabil karena digunakan sebagai base blend oleh roastery skala besar di seluruh dunia.`,
	},

	// ── B4: CCTC — Perbandingan standar grading ───────────────────────────────

	{
		ID:     "cctc-sni-vs-sca",
		Source: "CCTC, Perbandingan Standar Grading Kopi",
		Title:  "Perbedaan SNI vs SCA Grading System",
		Text: `Perbedaan utama grading SNI (Indonesia) vs SCA (Specialty Coffee Association, internasional): SNI berbasis nilai cacat fisik pada 300g sampel; SCA berbasis cupping score (skala 0–100) ditambah pemeriksaan cacat. Untuk SCA specialty: skor cupping minimal 80, primary defect (cacat kategori 1) = 0, secondary defect (kategori 2) maksimal 5 per 350g. SNI Grade 1 (cacat ≤11) umumnya berkorelasi dengan SCA specialty jika cupping score-nya juga ≥80. Keduanya saling melengkapi — SNI mengukur fisik, SCA mengukur rasa.`,
	},
	{
		ID:     "cctc-cupping-protocol",
		Source: "CCTC / SCA Cupping Protocol",
		Title:  "Protokol Cupping SCA dan Interpretasi Skor",
		Text: `Cupping score SCA menilai 10 atribut: Fragrance/Aroma, Flavor, Aftertaste, Acidity, Body, Balance, Uniformity, Clean Cup, Sweetness, dan Overall — dikurangi Defects. Total skor: di bawah 70 = tidak layak specialty; 70–79 = komersial; 80–84 = specialty; 85–89 = specialty premium; 90+ = exceptional/rare. Cupping dilakukan oleh Q-Grader bersertifikat CQI. Satu primary defect langsung mendiskualifikasi dari specialty. Skor ini dipakai sebagai acuan pembelian utama oleh roastery specialty.`,
	},
	{
		ID:     "cctc-primary-secondary-defects",
		Source: "CCTC / SCA Green Coffee Defects",
		Title:  "Primary vs Secondary Defects (SCA)",
		Text: `SCA membagi cacat kopi menjadi dua kategori: Primary defects (satu biji sudah gugurkan specialty) antara lain biji hitam penuh, biji busuk, biji berjamur, benda asing (kerikil, ranting), dan biji berkulit kopi (hull/husk). Secondary defects (batas 5 cacat per 350g) antara lain biji pecah sebagian, biji kulit ari (parchment), biji belum matang (unripe/quaker), biji berlubang serangga, dan biji coklat. Cara hitung berbeda dari SNI tapi konsepnya serupa.`,
	},

	// ── B5: Karakteristik origin kopi Indonesia ───────────────────────────────

	{
		ID:     "origin-gayo-aceh",
		Source: "Penelitian NCBI, Pengetahuan Specialty Coffee",
		Title:  "Profil Kopi Gayo, Aceh",
		Text: `Arabika Gayo dari Aceh (Bener Meriah dan Aceh Tengah) ditanam di ketinggian 1.200–1.500 mdpl. Profil rasa khas: herbal, earthy, dark chocolate, sedikit fruity, aftertaste panjang, body tebal. Asam medium-rendah dengan kompleksitas tinggi. Diproses mayoritas dengan giling basah (wet-hulled), sebagian kecil fully-washed atau natural untuk pasar specialty. Supply besar — termasuk origin arabika terbesar di Indonesia. Gayo memiliki Indikasi Geografis (IG) terdaftar.`,
	},
	{
		ID:     "origin-toraja-sulsel",
		Source: "Pengetahuan Specialty Coffee Indonesia",
		Title:  "Profil Kopi Toraja, Sulawesi Selatan",
		Text: `Arabika Toraja dari Tana Toraja dan Enrekang, Sulawesi Selatan, ditanam di 1.000–1.500 mdpl. Profil rasa: earthy, wine-like, dark chocolate, herbal, body sangat tebal. Permintaan sangat tinggi di Jepang — Toraja adalah salah satu arabika Indonesia paling dikenal di Asia Timur. Proses giling basah dominan. Beberapa lot honey process menghasilkan fruity notes lebih menonjol. Supply terbatas dibanding Gayo sehingga harga sering lebih tinggi.`,
	},
	{
		ID:     "origin-wamena-papua",
		Source: "Pengetahuan Specialty Coffee Indonesia",
		Title:  "Profil Kopi Wamena, Papua",
		Text: `Arabika Wamena dari Lembah Baliem, Papua, ditanam organik oleh komunitas lokal di ketinggian 1.500–2.000 mdpl. Profil: floral, fruity, sangat bersih (clean cup), aftertaste manis — salah satu arabika terbersih Indonesia. Produksi sangat terbatas (ratusan ton per tahun) sehingga harganya premium, bisa Rp 100.000–150.000/kg green bean. Supply ketat — harga naik tajam saat permintaan meningkat. Cocok untuk roastery yang mau menawarkan produk ultra-premium.`,
	},
	{
		ID:     "origin-java-preanger",
		Source: "Pengetahuan Specialty Coffee Indonesia",
		Title:  "Profil Kopi Java Preanger, Jawa Barat",
		Text: `Arabika Java Preanger dari Pengalengan, Garut, Ciwidey (Jawa Barat), ditanam 1.000–1.800 mdpl. Profil: bersih (clean), nutty, coklat susu, sedikit citrus, body medium. Proses umumnya fully-washed — menghasilkan kopi lebih bersih dan mudah diprediksi dibanding giling basah Sumatera. Dikenal di Eropa terutama Belanda (warisan perkebunan kolonial era VOC). Java grade 1 dengan skor SCA ≥80 dipasarkan sebagai Java specialty.`,
	},
	{
		ID:     "origin-flores-bajawa",
		Source: "Pengetahuan Specialty Coffee Indonesia",
		Title:  "Profil Kopi Flores Bajawa",
		Text: `Arabika Bajawa dari Ngada, Flores, NTT, ditanam di 1.000–1.800 mdpl. Profil: floral kuat, fruity (berry, citrus), caramel, dark chocolate, aftertaste panjang. Diproses semi-washed atau fully-washed. Media specialty coffee internasional mulai banyak menyebut Bajawa sebagai "hidden gem" Indonesia — demand meningkat beberapa tahun terakhir. Supply masih terbatas, kenaikan demand bisa membuat harga naik signifikan.`,
	},
	{
		ID:     "origin-mandailing-sumut",
		Source: "Pengetahuan Specialty Coffee Indonesia",
		Title:  "Profil Kopi Mandailing (Mandheling), Sumatera Utara",
		Text: `Arabika Mandailing dari Kabupaten Mandailing Natal, Sumatera Utara. Di pasar internasional dijual sebagai "Sumatran Mandheling". Profil: earthy, herbal, body sangat tebal, dark chocolate, asam rendah — karakter khas giling basah Sumatera. Populer di Amerika Serikat dan Eropa. Harga lebih terjangkau dari Gayo tapi tetap premium dibanding robusta. Volume produksi cukup besar, supply lebih stabil dari Wamena atau Flores.`,
	},
	{
		ID:     "origin-kintamani-bali",
		Source: "Pengetahuan Specialty Coffee Indonesia",
		Title:  "Profil Kopi Kintamani, Bali",
		Text: `Arabika Kintamani dari kawasan Danau Batur, Bali (900–1.500 mdpl). Profil: citrusy (lemon, jeruk), fruity, clean, body medium, brightness tinggi. Proses umumnya fully-washed karena pengaruh sistem pertanian organik lokal (subak). Memiliki Indikasi Geografis (IG) terdaftar. Pasar kuat di Bali (pariwisata), Jepang, dan Australia. Kecenderungan cocok untuk metode seduh V60, pour over, dan cold brew.`,
	},
	{
		ID:     "origin-lampung-robusta",
		Source: "AEKI-AICE / Pengetahuan Industri",
		Title:  "Profil Kopi Robusta Lampung",
		Text: `Robusta Lampung adalah kopi volume terbesar Indonesia. Profil: pahit kuat, earthy, cocoa, aftertaste panjang dan tebal. Kadar kafein lebih tinggi dari arabika. Harga jauh lebih rendah (Rp 35.000–50.000/kg) namun volume sangat besar. Ideal untuk blend yang membutuhkan body kuat dan efisiensi biaya. Digunakan roastery besar dan produsen kopi instan global. Robusta Temanggung (Jawa Tengah) lebih premium dari Lampung karena ketinggian lebih tinggi.`,
	},
	{
		ID:     "origin-sidikalang-sumut",
		Source: "Pengetahuan Specialty Coffee Indonesia",
		Title:  "Profil Kopi Sidikalang, Sumatera Utara",
		Text: `Arabika Sidikalang dari Kabupaten Dairi, Sumatera Utara, sekitar 1.000–1.400 mdpl. Termasuk dalam rumpun kopi Sumatera Utara bersama Mandailing dan Lintong. Profil: earthy, herbal, body tebal, dark fruit, asam medium. Diproses dengan giling basah. Harga lebih terjangkau dari Gayo atau Toraja, sehingga cocok untuk roastery yang ingin menawarkan arabika Sumatera kelas menengah dengan margin yang cukup.`,
	},

	// ── Proses kopi dan pengaruhnya ───────────────────────────────────────────

	{
		ID:     "processing-wet-hulled",
		Source: "Penelitian NCBI / Pengetahuan Proses Kopi",
		Title:  "Giling Basah (Wet-Hulled / Giling Basah Process)",
		Text: `Giling basah (wet-hulling atau Giling Basah) adalah metode pengolahan kopi khas Indonesia, terutama Sumatera. Biji dikupas saat masih basah (kadar air 30–40%), dikeringkan sebagian, lalu digiling untuk membuang kulit parchment. Hasilnya: biji berwarna hijau kebiruan dengan kadar air lebih tinggi, body sangat tebal, earthy/herbal yang kuat, asam rendah. Metode ini efisien untuk daerah lembap tapi menghasilkan profil rasa yang khas dan tidak cocok untuk semua pasar.`,
	},
	{
		ID:     "processing-washed",
		Source: "Pengetahuan Proses Kopi Specialty",
		Title:  "Fully-Washed Process",
		Text: `Fully-washed (wet process) adalah metode pengolahan kopi yang menghasilkan rasa bersih dan asam terang. Buah kopi dikupas, difermentasi dalam air 12–48 jam untuk melarutkan lapisan mucilage, dicuci bersih, lalu dikeringkan. Hasilnya: kopi lebih bersih, asam lebih cerah, fruity notes lebih jelas, cocok untuk pour over dan filter. Digunakan di Kintamani, Java Preanger, sebagian Flores. Lebih mudah dikontrol kualitasnya dibanding giling basah.`,
	},
	{
		ID:     "processing-natural-honey",
		Source: "Pengetahuan Proses Kopi Specialty",
		Title:  "Natural dan Honey Process",
		Text: `Natural process: buah kopi dikeringkan utuh tanpa dikupas. Menghasilkan rasa fruity intens, wine-like, body tebal, tapi lebih berisiko fermentasi tidak terkontrol. Honey process: kupas kulit luar, biarkan sebagian mucilage, keringkan. Hasilnya antara washed dan natural — fruity medium, bersih, sweet. Di Indonesia, natural dan honey mulai populer untuk lot-lot premium Gayo, Toraja, dan Flores yang ditujukan pasar specialty tinggi.`,
	},

	// ── Konteks harga dan bisnis ─────────────────────────────────────────────

	{
		ID:     "price-reference",
		Source: "Referensi Pasar / World Bank Pink Sheet 2025",
		Title:  "Referensi Harga Kopi Indonesia (Green Bean)",
		Text: `Harga green bean kopi Indonesia di tingkat pedagang lokal (perkiraan 2024–2025): Arabika specialty grade 1 (Gayo, Toraja, Wamena): Rp 85.000–150.000/kg; Arabika komersial grade 2–3 (Mandailing, Sidikalang): Rp 65.000–85.000/kg; Robusta premium (Temanggung): Rp 48.000–60.000/kg; Robusta komersial (Lampung): Rp 35.000–50.000/kg. Harga channel petani (farmer) biasanya 15–25% lebih murah dari middleman/tengkulak karena tidak ada markup perantara.`,
	},
	{
		ID:     "channel-comparison",
		Source: "Pengetahuan Industri Kopi Indonesia",
		Title:  "Perbandingan Channel Petani vs Tengkulak",
		Text: `Dua channel pengadaan biji kopi utama di Indonesia: (1) Petani langsung (farmer channel) — harga 15–25% lebih murah, kualitas lebih konsisten jika sudah ada hubungan, memerlukan minimum order biasanya 50–200 kg, butuh relationship jangka panjang dan kadang biaya logistik lebih tinggi; (2) Tengkulak/middleman — lebih fleksibel volume (bisa beli mulai 10–20 kg), mudah, tapi ada markup 15–30%. Untuk roastery kecil, tengkulak lebih praktis; untuk roastery yang fokus traceability dan volume menengah-besar, langsung petani lebih menguntungkan.`,
	},
	{
		ID:     "harvest-calendar",
		Source: "Pengetahuan Pertanian Kopi Indonesia",
		Title:  "Kalender Panen Kopi per Wilayah Indonesia",
		Text: `Kalender panen kopi Indonesia berdasarkan wilayah: Sumatera (Gayo, Mandailing, Sidikalang) — panen utama April–Juni, panen kecil Oktober–Desember; Jawa (Preanger, Temanggung) — Agustus–Oktober; Sulawesi (Toraja, Enrekang) — Juni–September; Flores (Bajawa) — Juni–Agustus; Bali (Kintamani) — Juni–September; Papua (Wamena) — Oktober–Desember. Membeli sedekat mungkin setelah panen memastikan biji segar. Stok yang sudah 12+ bulan pasca-panen mengalami degradasi rasa signifikan.`,
	},
	{
		ID:     "storage-best-practices",
		Source: "SNI 01-2907-2008 / Pengetahuan Industri",
		Title:  "Penyimpanan Green Bean yang Baik",
		Text: `Best practice penyimpanan green bean: (1) Suhu gudang 18–22°C; (2) Kelembapan relatif gudang (RH) dijaga 50–70%RH; (3) Kadar air biji maksimal 12,5% (SNI) — berbeda dari RH gudang; (4) Gunakan karung GrainPro atau vakum untuk lot specialty; (5) Jauhkan dari bau menyengat (cat, solar, rempah); (6) Putar stok FIFO — biji lama 12 bulan mulai terasa "past crop" dan degradasi asam; (7) Cek moisture tiap 3 bulan. Kondisi penyimpanan buruk bisa menurunkan kopi grade 1 menjadi grade 2–3 dalam beberapa bulan.`,
	},
	{
		ID:     "margin-analysis",
		Source: "Pengetahuan Bisnis Roastery",
		Title:  "Analisis Margin Roastery Kopi",
		Text: `Margin roastery kopi Indonesia: Harga beli green bean arabika specialty Rp 85.000–130.000/kg; setelah roasting, weight loss sekitar 15–20% (menyusut); harga jual roasted bean specialty Rp 130.000–200.000/kg. Margin kasar per kg roasted: Rp 30.000–70.000 tergantung origin dan grade. Robusta: beli Rp 40.000/kg, jual roasted Rp 70.000–85.000/kg, margin lebih tipis tapi volume lebih tinggi. Biaya roasting (listrik, tenaga, kemasan) biasanya Rp 10.000–20.000/kg.`,
	},
	{
		ID:     "crm-followup-patterns",
		Source: "Pengetahuan Bisnis Roastery",
		Title:  "Pola Follow-up Pelanggan Kafe",
		Text: `Pola pemesanan ulang kafe yang sehat: interval 7–14 hari untuk kafe aktif, 14–30 hari untuk kafe kecil. Tanda-tanda perlu follow-up segera: (1) interval melambat dari biasanya — mungkin ada masalah stok, keuangan, atau beralih supplier; (2) status "overdue" atau "credit" lebih dari 30 hari; (3) tidak ada respons setelah dihubungi. Rekomendasi: hubungi via WhatsApp dalam 1–2 hari setelah jatuh tempo. Pesan yang efektif: personal, menyebut nama, mengingatkan riwayat order spesifik, menawarkan solusi (cicilan, order lebih kecil).`,
	},
	{
		ID:     "specialty-trend-indonesia",
		Source: "Outlook Kopi 2023 / Tren Pasar",
		Title:  "Tren Specialty Coffee Indonesia 2022–2025",
		Text: `Tren specialty coffee Indonesia: (1) Pertumbuhan kafe specialty 15–20% per tahun di kota besar (Jakarta, Bandung, Surabaya, Bali); (2) Konsumen semakin literasi — single origin, traceability, metode seduh pour over dan AeroPress diminati; (3) Arabika Gayo dan Flores jadi favorit; (4) Roastery kecil dan menengah bertumbuh pesat, butuh green bean graded konsisten; (5) Ekspor specialty naik dengan harga premium ke Jepang, Korea, dan Australia. Roastery yang bisa menjamin grade konsisten dan traceability punya keunggulan kompetitif.`,
	},
}
