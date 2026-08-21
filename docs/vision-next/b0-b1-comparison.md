# Perbandingan Konkret Engine B0 versus B1

## Tujuan adapter

Adapter bukan untuk membuat B0 tampak seperti B1. Adapter hanya mengubah keluaran dua engine ke **kontrak evaluasi yang sama**, sehingga evaluator dapat menjawab:

1. Apa yang dideteksi masing-masing engine?
2. Seberapa stabil region/relasinya?
3. Apa yang hanya diketahui B0 karena classifier UI?
4. Apa yang ditemukan B1 sebagai evidence/geometry baru?
5. Berapa biaya CPU dan memory masing-masing?
6. Apakah peningkatan B1 benar-benar grounded atau hanya menghasilkan lebih banyak output?

B0 tetap berjalan menggunakan `facts.Extract` lama. B1 berjalan menggunakan canonical view, evidence field, region proposal, hypothesis, dan relation engine baru. Tidak ada output B0 yang dimasukkan ke B1 sebagai input, dan tidak ada parameter B1 yang disuntikkan ke B0.

## Kontrak normalisasi

Kedua engine dikonversi menjadi `eval.Observation`:

```go
type Observation struct {
    Engine      string            `json:"engine"`
    ImagePath   string            `json:"image_path"`
    Width       int               `json:"width"`
    Height      int               `json:"height"`
    ElapsedMS   int64             `json:"elapsed_ms"`
    Graph       schema.SceneGraph `json:"graph"`
    LegacyBoxes int               `json:"legacy_boxes,omitempty"`
}
```

Struktur graph bersama berisi `Region`, `Hypothesis`, `Relation`, `EvidenceRef`, `Warning`, dan `Provenance`. Karena B0 tidak memiliki evidence field baru, adapter B0 mencatat sumbernya sebagai `legacy-box` atau `legacy-classification`. Itu penting: **ketiadaan evidence B0 terlihat sebagai ketiadaan evidence**, bukan dipalsukan menjadi evidence B1.

## Jalur B0

Adapter B0 memanggil engine lama secara langsung:

```text
RunB0(path, options)
  → facts.Extract(path, mode, version, legacyConfig)
  → result.Boxes
  → setiap Box menjadi Region
  → setiap Box.Type menjadi Hypothesis.Label
  → Box.Score menjadi Hypothesis.Score
  → relation.Build(region boxes)
  → graph.Build(..., provenance.engine_version="b0-legacy")
  → Observation{Engine:"B0", ...}
```

Pemetaan box B0 dilakukan seperti berikut:

| Field B0 | Field normalisasi | Perlakuan |
|---|---|---|
| `x,y,w,h` | `Region.BBox` | Dipertahankan pada koordinat output B0 |
| `w*h` | `Region.Area` | Dihitung ulang |
| `type` | `Hypothesis.Label` | Misalnya `button`, `card`, `input` |
| `score` | `Hypothesis.Score` | Dipertahankan, lalu dibatasi ke `[0,1]` |
| `color`, `text`, `caption` | `EvidenceRef.Note`/metadata lanjutan | Tidak dijadikan bukti visual B1 |
| `order`, `page_type` | Metadata terpisah | Tidak dicampur dengan graph node score |
| `elapsed_ms` | `Observation.ElapsedMS` | Dipakai untuk runtime comparison |

B0 tetap dapat menyebut `button` atau `card`, tetapi provenance-nya akan menunjukkan `stage: b0-detector`. Ini membedakan **label classifier lama** dari **evidence region baru**.

## Jalur B1

Adapter B1 tidak memanggil `facts.Extract`. Ia menjalankan pipeline baru:

```text
RunB1(path, options)
  → processor.Decode(path)
  → imageview.FromImage(image, maxSide)
  → evidence.Compute(view)
  → region.Propose(field, config)
  → hypothesis.Generate(regions, width, height)
  → relation.Build(regions)
  → graph.Build(regions, hypotheses, relations, provenance)
  → Observation{Engine:"B1", ...}
```

Pada tahap B1 awal, hypothesis bersifat **shape-neutral**, misalnya `region`, `elongated_region`, `compact_region`, atau `flat_surface_region`. B1 belum boleh menyebut `person`, `car`, atau `cup` sebelum semantic prior dan benchmark label masuk. Ini sengaja agar B1 tidak mengarang pemahaman semantik sebelum evidence-nya tersedia.

## Koordinat dan fairness

B0 dan B1 harus dibandingkan pada coordinate frame yang didefinisikan jelas. B0 saat ini dapat resize ke `max_side` lalu mengembalikan koordinat pada gambar hasil resize. B1 juga menggunakan `max_side`, sehingga evaluator harus menyimpan `width` dan `height` setiap observation.

Untuk metric antar-engine, semua box/region diproyeksikan kembali ke **canonical evaluation canvas**:

```text
x_eval = x * W_eval / W_engine
 y_eval = y * H_eval / H_engine
 w_eval = w * W_eval / W_engine
 h_eval = h * H_eval / H_engine
```

Jangan membandingkan koordinat mentah bila ukuran output berbeda. Untuk latency, bagaimanapun, resize dan preprocessing masing-masing engine harus dihitung sebagai bagian dari end-to-end runtime.

## Matching B0 dan B1

Matching bukan berdasarkan urutan node. Urutan dapat berubah dan tidak merepresentasikan identitas objek. Gunakan greedy matching atau Hungarian assignment dengan cost:

```text
cost(a,b) = 1 - IoU(a.bbox, b.bbox)
           + λ_type · type_penalty(a,b)
           + λ_size · size_penalty(a,b)
```

Untuk perbandingan pertama, gunakan dua mode:

| Mode | Matching memakai | Tujuan |
|---|---|---|
| Geometry-only | IoU dan center distance | Perbandingan region tanpa bias label |
| Label-aware | Geometry + label compatibility | Perbandingan klasifikasi saat B1 sudah punya semantic prior |

Geometry-only wajib menjadi angka utama di awal. Kalau B0 mempunyai label UI dan B1 baru hanya mempunyai shape hypothesis, label-aware tidak adil.

Contoh:

```text
B0: button [50,29,101,32]
B1: elongated_region [48,28,105,35]
IoU = 0.78

Kesimpulan:
B1 berhasil menemukan region geometris yang sama,
bukan “gagal” hanya karena belum menamai region sebagai button.
```

## Metrik konkret

### 1. Coverage dan geometry agreement

Untuk setiap region B0 yang memiliki pasangan B1 pada IoU ≥ 0,5:

```text
coverage_B1_on_B0 = matched_B0_regions / total_B0_regions
```

Untuk setiap pair matched, simpan IoU, center distance ter-normalisasi, rasio luas, dan boundary overlap. Ini menjawab apakah B1 mempertahankan struktur yang sudah dikenal B0.

### 2. Novel discovery

Region B1 yang tidak memiliki pasangan B0 disebut `B1_novel`. Namun novel bukan otomatis benar. Ia harus diuji terhadap annotation atau human review:

```text
novel_precision = validated_novel_B1 / all_novel_B1
novel_recall    = validated_novel_B1 / all_ground_truth_regions_not_found_by_B0
```

Tanpa validasi, “B1 menemukan lebih banyak” hanya berarti B1 mengeluarkan lebih banyak region.

### 3. Duplicate/fragmentation

B1 dapat memecah satu objek menjadi banyak region. Ukur:

```text
fragmentation = matched_B1_regions_per_ground_truth_object
merge_error   = ground_truth_objects_per_B1_region
```

Region count yang lebih tinggi tidak dianggap lebih canggih bila fragmentation meningkat tanpa evidence quality.

### 4. Relation agreement

Bandingkan edge visible yang dapat dihitung dari geometry:

```text
relation_precision = matched_correct_edges / all_B1_edges
relation_recall    = matched_correct_edges / all_reference_edges
```

Edge `inferred` seperti `holding` tidak boleh dihitung dalam visible geometry score. Ia masuk track terpisah dengan label dan evidence-nya sendiri.

### 5. Provenance dan uncertainty

Hitung berapa persen hypothesis/edge yang memiliki `EvidenceRef`, serta apakah confidence turun pada gambar transformasi/ambigu:

```text
provenance_coverage = claims_with_evidence / total_claims
unsupported_rate    = claims_without_valid_support / total_claims
```

B1 boleh memiliki accuracy lebih rendah pada fase awal, tetapi unsupported rate dan calibration harus lebih baik daripada sistem yang memaksa label.

### 6. Runtime

Catat minimal:

```text
cold_start_ms
warm_p50_ms
warm_p95_ms
peak_rss_mb
threads
input_pixels
regions_per_ms
```

B0 dan B1 dijalankan dalam proses terpisah atau worker yang di-reset, agar cache/global state satu engine tidak menguntungkan engine lain secara tidak adil.

## Alur evaluator konkret

```text
manifest.json
  ├── sample_001.png
  ├── sample_002.png
  └── ...

for each sample:
  run B0 in isolated measurement
  run B1 in isolated measurement
  normalize both to Observation
  save predictions.jsonl
  match regions geometry-only
  compare relations and provenance
  if annotations exist:
      score region/label/relation/grounding
  save latency.csv and failures.jsonl
aggregate metrics per domain and globally
```

Satu baris `predictions.jsonl` dapat berbentuk:

```json
{
  "sample_id": "photo_001",
  "b0": {
    "engine": "B0",
    "width": 1024,
    "height": 768,
    "regions": 7,
    "relations": 10,
    "elapsed_ms": 31,
    "provenance_coverage": 1.0
  },
  "b1": {
    "engine": "B1",
    "width": 1024,
    "height": 768,
    "regions": 13,
    "relations": 18,
    "elapsed_ms": 44,
    "provenance_coverage": 1.0
  },
  "geometry": {
    "b0_b1_matched": 5,
    "mean_iou": 0.71,
    "b1_novel": 8,
    "b1_novel_validated": 5,
    "fragmentation": 1.4
  },
  "status": "ok"
}
```

Angka di atas hanya contoh format, bukan hasil benchmark.

## Apa yang dianggap kemenangan B1?

B1 dianggap lebih baik bukan hanya ketika jumlah region atau label lebih banyak. Kemenangan minimal harus berupa salah satu dari kondisi berikut tanpa merusak kondisi lain:

| Kemenangan | Bukti |
|---|---|
| Lebih stabil | Geometry/evidence konsisten terhadap transformasi gambar |
| Lebih grounded | Unsupported claim rate lebih rendah dan provenance coverage lebih tinggi |
| Lebih luas | Menemukan region/relasi di foto, dokumen, diagram, dan screenshot, bukan UI saja |
| Lebih berguna | Query graph compositional benar pada subset terdefinisi |
| Lebih efisien | Quality-per-ms atau quality-per-MB lebih baik pada CPU yang sama |
| Lebih jujur | Abstention/calibration lebih baik pada gambar ambigu |

B1 tidak dianggap menang bila hanya mengeluarkan lebih banyak region, lebih banyak edge, atau caption yang lebih panjang tanpa peningkatan metric di atas.

## Status implementasi saat ini

Adapter B0 dan B1 awal sudah dibuat di `internal/visionnext/eval/adapter.go`. `RunB0` memanggil `facts.Extract` lama lalu menormalisasi box ke graph. `RunB1` menjalankan canonical view, evidence, region, hypothesis, relation, dan graph builder baru. Integration test untuk kedua jalur telah lulus pada fixture repository.

Langkah berikutnya adalah menambahkan matcher IoU, evaluator JSONL, serta output per-domain. Setelah itu baru kita bisa menjawab secara numerik apakah B1 benar-benar lebih baik daripada B0.
