# Semantic Fixture Kecil dan Logika CPU-First HeraVision Next

## Kesimpulan singkat

Kita **bisa membuat semantic layer ringan yang cukup pintar untuk kategori terbatas**, tetapi tidak realistis mengharapkan matematika sederhana memahami semua gambar, semua ekspresi, semua konteks, dan semua aktivitas seperti model vision besar pada PC kentang. Batasnya bukan hanya kecepatan komputer. Sebagian informasi memang tidak tersedia atau terlalu ambigu di pixel.

Solusi paling kuat adalah **conditional computation**: mesin tidak menjalankan semua perhitungan pada semua gambar. Ia mulai dari bukti murah, naik ke analisis lebih mahal hanya jika diperlukan, dan abstain bila informasi yang tersedia tidak cukup.

> Target realistis HeraVision bukan “selalu tahu”. Targetnya adalah “tahu kapan tahu, tahu kapan tidak tahu, dan setiap jawaban punya bukti.”

## 1. Apa itu semantic fixture 30–60 gambar?

Semantic fixture adalah **kumpulan pengujian**, bukan seluruh dataset training. Ia dipakai untuk mengukur apakah engine dapat menghubungkan struktur visual dengan kategori semantic sederhana. Setiap gambar harus memiliki label, region pendukung, status answerability, dan provenance. Gambar tidak boleh dibuat secara acak atau diberi label berdasarkan tebakan mesin.

Untuk tahap pertama, saya menyarankan 48 gambar, lalu menambah menjadi 60 setelah protokolnya stabil.

| Domain | Jumlah awal | Contoh kategori yang diuji |
|---|---:|---|
| Foto objek | 8 | object, vehicle, tool, appliance |
| Hewan | 8 | animal, single-animal, multi-animal |
| Manusia | 8 | person-present, person-group, person-with-object |
| Ekspresi wajah | 8 | face-visible, face-occluded, coarse-expression-possible |
| Dokumen/screenshot | 8 | document, table-like, interface-like, text-dense |
| Diagram | 8 | flowchart, network, schematic, chart |
| **Total awal** | **48** | — |

Empat puluh delapan gambar tersebut bukan berarti 48 gambar acak. Setiap kategori harus memiliki contoh mudah, contoh sulit, dan contoh negatif yang mirip. Misalnya kategori `animal` tidak cukup berisi delapan foto anjing yang jelas; harus ada satu hewan kecil, dua hewan, hewan tertutup, hewan di background, dan gambar non-hewan yang memiliki tekstur mirip.

### Tahap 30 gambar

Tahap pertama dapat memakai 30 fixture yang sudah ada sebagai **structural-semantic pilot**, tetapi hasilnya hanya menguji kategori luas seperti `person-present`, `animal-present`, `diagram`, `object`, dan `insufficient-evidence`. Jangan langsung menguji emosi detail pada 30 gambar ini karena sebagian besar tidak dirancang sebagai dataset wajah.

### Tahap 48–60 gambar

Tambahan gambar harus berasal dari data nyata. Untuk foto umum, kita dapat memilih subset Imagenette/Imagewoof dan data Wikimedia yang sudah tersedia. Untuk ekspresi wajah, kita membutuhkan subset nyata yang memang memiliki wajah cukup besar dan label ekspresi yang sah. Jika dataset ekspresi belum tersedia lokal, saya tidak akan mengarang label; kita harus mengambil dataset publik yang lisensinya jelas atau meminta Anda menyediakan data.

## 2. Label yang aman untuk tahap awal

Kita harus memisahkan tiga tingkat semantic. Tingkat pertama hanya menjawab **kehadiran dan struktur**, tingkat kedua menjawab kategori luas, dan tingkat ketiga baru menjawab atribut atau ekspresi.

| Level | Contoh label | Status pengembangan |
|---|---|---|
| L0: visual structure | region, boundary, group, diagram-like | Sudah didukung B1 |
| L1: broad semantic | animal, person, vehicle, document, diagram | Kandidat berikutnya |
| L2: attributes | single/multiple, occluded, small/large, indoor/outdoor | Setelah L1 stabil |
| L3: expression/activity | smiling, angry, running, holding, driving | Belum boleh diklaim |

Kategori luas lebih cocok untuk CPU dan lebih mudah diaudit. Label `animal` masih dapat didukung oleh evidence bentuk dan region. Label `angry` membutuhkan wajah, kualitas crop, landmark, konteks, dan label manusia yang konsisten. Dua gambar dengan wajah sama dapat tampak berbeda karena cahaya, sudut, budaya, atau konteks; karena itu ekspresi harus menjadi micro-benchmark tersendiri.

## 3. Format satu sample

Setiap sample semantic sebaiknya memiliki struktur seperti berikut.

```json
{
  "id": "semantic-md2-person-fish-001",
  "image_path": "data/.../image.jpg",
  "domain": "natural-photo",
  "sha256": "...",
  "labels": ["person", "animal"],
  "attributes": ["multi-foreground", "held-object"],
  "answerability": "answerable-broad-semantic",
  "regions": [
    {
      "id": "gt-person",
      "bbox": {"x": 37, "y": 12, "w": 140, "h": 134},
      "role": "semantic-support",
      "visibility": "visible"
    }
  ],
  "required_evidence": ["region", "boundary", "scale-stability"],
  "review_status": "internal-development-review"
}
```

`labels` adalah target evaluasi. `regions` menjelaskan di mana evidence berada. `answerability` mengatakan apakah klaim broad semantic aman untuk gambar tersebut. `required_evidence` mencegah classifier menang hanya karena warna global tanpa region yang dapat diaudit.

## 4. Bagaimana label ekspresi dibuat?

Ekspresi tidak boleh dimasukkan sebagai label tambahan pada foto umum secara asal. Untuk micro-benchmark ekspresi, setiap gambar harus memiliki:

| Komponen | Alasan |
|---|---|
| face bbox | Memastikan mesin tidak membaca seluruh gambar sebagai wajah |
| face quality | Menilai ukuran, blur, occlusion, dan sudut wajah |
| coarse label | Misalnya neutral, positive, negative, atau uncertain |
| annotator agreement | Ekspresi sering subjektif |
| abstention label | Wajah kecil/tertutup harus boleh ditolak |

Saya menyarankan mulai dari `face-visible` dan `expression-uncertain` sebelum `happy`, `sad`, atau `angry`. Kalau dua reviewer manusia tidak sepakat, label harus menjadi `uncertain`, bukan dipaksa menjadi salah satu emosi.

Secara matematika, ekspresi adalah masalah sulit karena sinyalnya kecil. Perubahan pixel akibat ekspresi bisa lebih kecil daripada perubahan akibat kompresi, cahaya, pose, dan kamera. Sistem CPU-first dapat mengenali beberapa pola sederhana pada face crop yang besar, tetapi tidak boleh menjanjikan pemahaman semua ekspresi atau keadaan batin.

## 5. Logika matematika CPU-first

### 5.1 Feature vector berlapis

Dari setiap region, engine membentuk vektor fitur ringan:

```text
z = [
  area_ratio,
  aspect_ratio,
  boundary_strength,
  scale_stability,
  gradient_density,
  orientation_histogram,
  local_contrast,
  flatness,
  chroma_magnitude,
  texture_summary,
  position_x,
  position_y,
  neighbor_count
]
```

Fitur ini bukan label. Fitur hanya menggambarkan bukti. Untuk semantic broad, kita dapat menambah histogram warna sederhana, moments bentuk, jumlah region anak, dan pola hubungan. Tidak perlu menyimpan seluruh pixel dalam model classifier.

### 5.2 Classifier kecil

Untuk label `animal`, `person`, `vehicle`, atau `diagram`, kandidat awal dapat memakai logistic scorer:

```text
s_k(z) = w_k · z + b_k
p_k(z) = 1 / (1 + exp(-s_k(z)))
```

`w_k` dan `b_k` adalah bobot kecil. Inference hanya membutuhkan dot product, bukan transformer besar. Model tetap harus dilatih pada development split dan diuji pada holdout domain.

### 5.3 Evidence-aware score

Probabilitas classifier tidak boleh langsung menjadi jawaban. Kita gabungkan skor semantic dengan bukti visual:

```text
E_k = p_k × R × B × S × Q
```

Dengan:

- `R` = kualitas region, misalnya region valid dan tidak terlalu terfragmentasi;
- `B` = boundary evidence;
- `S` = scale stability;
- `Q` = quality input, termasuk ukuran crop, blur, dan occlusion.

Jika salah satu faktor sangat rendah, `E_k` turun. Keputusan harus menggunakan dua syarat:

```text
answered jika E_k >= τ dan (E_k - E_second) >= δ
abstain jika region ada tetapi E_k < τ atau margin < δ
insufficient_evidence jika tidak ada region yang valid
```

`τ` adalah threshold keamanan. `δ` adalah margin agar mesin tidak menjawab ketika dua label sama-sama mungkin. Nilai ini tidak boleh dipilih dari feeling; harus diukur pada development fixture.

### 5.4 Logika multi-region

Satu label harus mempunyai dukungan region. Untuk beberapa region, kita dapat memakai agregasi probabilitas yang tidak menghitung bukti sama berulang kali:

```text
support(k) = 1 - Π_i (1 - p_k,i × q_i)
```

`q_i` adalah kualitas evidence region ke-i. Rumus ini membuat beberapa region independen dapat memperkuat dukungan, tetapi satu region buruk tidak otomatis membuat jawaban sangat yakin. Jika region saling tumpang tindih tinggi, faktor redundansi harus mengurangi kontribusinya.

### 5.5 Relasi dan ekspresi

Relasi seperti `above`, `left_of`, dan `contains` dapat dihitung dari geometry. `touching` memerlukan boundary evidence. Ekspresi membutuhkan face crop dan quality gate:

```text
expression_answerable = face_size × face_quality × landmark_stability >= τ_face
```

Jika wajah terlalu kecil atau tertutup, mesin harus menjawab `expression-uncertain`, bukan menebak emosi.

## 6. Apakah ini berat di PC kentang?

Dengan feature extraction dan logistic scorer kecil, semantic layer broad **seharusnya ringan**, tetapi angka final harus diukur. Beban utama bukan dot product; beban utamanya adalah decode gambar, multi-scale proposal, OCR, dan analisis crop.

| Mode | Komponen | Perkiraan engineering |
|---|---|---|
| Geometry | B1 region + relation | Paling ringan, sudah berjalan CPU |
| Semantic-lite | Geometry + feature vector + logistic scorer | Tambahan kecil setelah region tersedia |
| Face-lite | Semantic-lite + face crop/landmark | Lebih mahal dan hanya dipanggil jika wajah terdeteksi |
| Full semantic | Banyak crop, OCR, model besar, caption | Tidak cocok untuk PC kentang tanpa optimasi berat |

Arsitektur cascade menghindari pemrosesan mahal:

```text
decode
→ cheap structural scan
→ region filter
→ broad semantic scorer
→ face/object specialist hanya jika diperlukan
→ abstain jika evidence tidak cukup
```

Dengan cara ini gambar diagram tidak perlu melewati face-expression pipeline, dan gambar tanpa wajah tidak membayar biaya face detector. Ini adalah penghematan paling nyata: **mengurangi jumlah kasus yang masuk ke tahap mahal**, bukan menemukan rumus ajaib yang menghapus kebutuhan komputasi.

## 7. Apakah kita bisa memangkas komputasi secara “visioner”?

Bisa melakukan inovasi arsitektur, tetapi harus jujur tentang batasnya. Tidak ada rumus yang dapat membuat informasi yang tidak ada di gambar menjadi tersedia. Jika wajah hanya 8 pixel, tidak ada logika CPU yang dapat memastikan ekspresi dengan benar. Jika dua objek tertutup total, sistem hanya dapat membuat hipotesis atau abstain.

Penemuan yang realistis dan dapat diuji adalah **evidence-driven sparse vision**:

1. Menggunakan region yang stabil sebagai unit komputasi, bukan seluruh pixel berulang kali.
2. Menggunakan event-driven cascade, sehingga specialist hanya aktif ketika evidence murah mengindikasikan kebutuhan.
3. Menggunakan multi-scale secara selektif pada region yang tidak stabil, bukan seluruh gambar pada semua skala.
4. Menggunakan graph sparsification untuk menghindari semua pasangan region.
5. Menggunakan early exit jika satu label sudah kuat dan margin terhadap label kedua besar.
6. Menggunakan abstention sebagai penghemat komputasi dan pencegah klaim salah.
7. Menyimpan provenance ringkas sehingga audit tidak perlu mengulang seluruh pipeline.

Inilah hipotesis penelitian HeraVision: **representasi visual berlapis yang sparse dan evidence-first dapat memberikan semantic broad yang berguna dengan CPU jauh lebih kecil daripada model vision umum, tetapi dengan coverage yang lebih rendah dan abstention yang lebih sering.** Hipotesis itu harus dibuktikan melalui benchmark, bukan slogan.

## 8. Eksperimen yang harus dilakukan

| Eksperimen | Yang dibandingkan | Gate |
|---|---|---|
| A | Geometry-only versus semantic-lite | Semantic tidak boleh menambah error graph |
| B | Threshold 0.45, 0.65, 0.70 | Unsafe-answer turun tanpa coverage runtuh |
| C | Specialist selalu aktif versus cascade | Runtime dan RAM turun dengan kualitas setara |
| D | Semua scale versus selective scale | Recall tidak jatuh melewati batas |
| E | Ada evidence provenance versus tanpa provenance | Tidak boleh ada answered tanpa region evidence |
| F | Wajah jelas versus wajah kecil/tertutup | Expression specialist harus abstain pada kualitas rendah |

Setiap eksperimen harus mencatat git SHA, hardware, GOMAXPROCS, memory limit, max-side, runtime, region count, answer status, dan error count.

## 9. Metrik yang benar

Accuracy saja tidak cukup. Untuk semantic CPU-first kita perlu mengukur:

| Metrik | Makna |
|---|---|
| Macro-F1 | Kualitas lintas kategori tanpa didominasi kelas besar |
| Coverage | Persentase sample yang dijawab |
| Selective accuracy | Akurasi hanya pada sample yang dijawab |
| Unsafe-answer rate | Jawaban positif yang salah pada sample ambigu |
| Correct-abstention rate | Seberapa sering sistem menolak kasus yang memang tidak aman |
| Evidence completeness | Persentase jawaban yang memiliki region dan provenance valid |
| p95 runtime | Latency tail pada PC kentang |
| Peak memory | Batas RAM sebenarnya |

Fixture 30–60 gambar terlalu kecil untuk klaim umum. Ia cocok sebagai **development smoke benchmark**. Untuk klaim serius, kita perlu holdout lintas domain yang lebih besar dan review independen.

## Keputusan engineering

Tahap semantic berikutnya sebaiknya tidak dimulai dari “HeraVision memahami semua ekspresi”. Tahap yang benar adalah:

```text
broad semantic fixture
→ evidence-aware logistic scorer
→ threshold dan margin
→ cascade specialist
→ face micro-benchmark
→ holdout multi-domain
→ semantic expansion
```

Jika eksperimen menunjukkan semantic-lite hanya menaikkan error dan tidak memberi manfaat, kita tidak memaksanya. Model baru hanya dipertahankan jika memberi peningkatan terukur dengan runtime dan memory tetap di budget.

> PC kentang dapat menjalankan mesin vision yang **sangat disiplin dan cukup pintar pada ruang lingkup terbatas**. Ia tidak dapat secara ajaib memperoleh pemahaman dunia yang tidak didukung pixel, data, atau model. Keunggulan HeraVision harus datang dari representasi sparse, cascade selektif, provenance, dan keberanian untuk abstain.
