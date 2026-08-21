# HeraVision Next — Rencana Implementasi

## Status awal

Branch `vision-next` dibuat dari `main` pada tag `v0.1.2`. Engine lama dipertahankan sebagai baseline B0; tidak boleh dihapus atau diubah sampai evaluator baru dapat mengukur perbedaannya. Prioritas binary kecil tidak lagi menjadi constraint utama. Constraint utama adalah kualitas grounded, reproducibility, CPU latency, memory, dan uncertainty.

## Prinsip pembangunan

Kita membangun dari **kontrak dan evaluator menuju algoritma**, bukan sebaliknya. Setiap komponen baru harus menghasilkan artefak yang dapat diuji. Tidak boleh langsung membuat “captioner” atau chatbot sebelum region, evidence, provenance, dan metric runner tersedia.

## Fase 1 — Kontrak data dan baseline branch

Buat package baru yang terpisah dari `internal/facts` lama:

```text
internal/visionnext/
  schema/       # structs dan JSON schema
  imageview/    # canonical views dan coordinate transforms
  evidence/    # feature fields
  region/      # region proposals dan masks
  hypothesis/  # object hypotheses
  relation/    # geometric relations
  graph/       # scene graph dan provenance
  reasoning/   # query planner dan abstention
  runtime/     # orchestration dan budgets
```

Kontrak minimum harus mencakup:

```go
type Region struct {
    ID         string    `json:"id"`
    BBox       Rect      `json:"bbox"`
    Area       int       `json:"area"`
    PolygonRef string    `json:"polygon_ref,omitempty"`
    Features   Features  `json:"features"`
    Evidence   []EvidenceRef `json:"evidence"`
}

type Hypothesis struct {
    RegionIDs []string   `json:"region_ids"`
    Label     string     `json:"label"`
    Score     float64    `json:"score"`
    Uncertainty float64  `json:"uncertainty"`
    Evidence  []EvidenceRef `json:"evidence"`
}

type Relation struct {
    From       string     `json:"from"`
    To         string     `json:"to"`
    Predicate  string     `json:"predicate"`
    Status     string     `json:"status"` // visible | inferred
    Score      float64    `json:"score"`
    Evidence   []EvidenceRef `json:"evidence"`
}

type SceneGraph struct {
    Nodes      []Node     `json:"nodes"`
    Edges      []Relation `json:"edges"`
    Warnings   []Warning  `json:"warnings"`
    Provenance Provenance `json:"provenance"`
}
```

`EvidenceRef` minimalnya menyimpan `kind`, koordinat, scale, numeric value, dan source stage. Jawaban apa pun di masa depan harus dapat menunjuk `EvidenceRef`.

**Selesai jika:** schema dapat di-marshal/unmarshal, mempunyai golden JSON, dan mendukung unknown/abstention tanpa memakai string error sebagai pengganti status.

## Fase 2 — Evaluator reproducible

Buat command terpisah, misalnya `heravision eval`, tanpa mengubah command lama. Inputnya adalah manifest gambar dan annotations; outputnya satu directory run:

```text
runs/<timestamp>-<git_sha>/
  config.json
  hardware.json
  manifest.json
  predictions.jsonl
  graphs.jsonl
  metrics.json
  latency.csv
  failures.jsonl
```

Evaluator harus dapat menjalankan B0, B1, dan versi berikutnya dengan manifest yang sama. Minimal metric runner:

| Kategori | Metrik awal |
|---|---|
| Region | IoU, boundary F-score |
| Label | macro-F1, Recall@K |
| Relation | precision, recall, graph edit distance |
| Grounding | evidence coverage, unsupported claim rate |
| Uncertainty | ECE, Brier, risk-coverage |
| Runtime | cold/warm p50/p95, peak RSS, thread count |

**Selesai jika:** satu command menghasilkan metrics dan failure examples dari fixture nyata; hasil dapat diulang dari commit dan config yang sama.

## Fase 3 — Perception core

Implementasi dilakukan bertahap dan setiap tahap memiliki ablation:

1. `imageview`: decode, EXIF, coordinate system, grayscale, log-chroma, Gaussian pyramid.
2. `evidence`: gradient, edge, corner, local contrast, flatness, color statistics, LBP/texture ringan.
3. `region`: over-segmentation konservatif, adjacency graph, region merge cost, mask/polygon compression.
4. `stability`: hitung konsistensi region pada beberapa scale dan transformasi.

Cost awal region merge:

```text
w(i,j) = α·Δluminance
       + β·Δchroma
       + γ·texture_distance
       + δ·edge_barrier
       + η·scale_inconsistency
```

Jangan langsung mengoptimalkan nilai `α..η` pada test set. Mulai dari config eksplisit, calibration split, lalu lakukan sensitivity sweep.

**Selesai jika:** B1 dapat mengeluarkan region dan evidence pada foto, dokumen, diagram, screenshot, dan gambar low-quality tanpa label semantic palsu.

## Fase 4 — Hypothesis dan scene graph

Hypothesis object dibentuk dari region tunggal atau gabungan region dengan feature shape, color, texture, boundary, symmetry, dan scale stability. Jangan memaksa satu label; simpan top-k hypothesis.

Relasi awal yang aman untuk dibangun tanpa prior bahasa:

```text
left_of, right_of, above, below,
inside, contains, touching, overlapping,
aligned_with, near, separated_from
```

Relasi seperti `holding`, `using`, atau `belongs_to` harus berstatus `inferred` dan hanya boleh muncul melalui rule/prior yang eksplisit. Scene graph harus memisahkan `visible` dari `inferred`.

**Selesai jika:** graph dapat divisualisasikan, setiap edge memiliki evidence, dan query seperti “apa yang berada di atas region meja?” dapat dijawab tanpa decoder bahasa besar.

## Fase 5 — Reasoning dan semantic prior opsional

Bangun query planner berbasis operasi graph:

```text
question constraints
  -> entity candidates
  -> relation traversal
  -> evidence filter
  -> confidence calibration
  -> answer/abstain
```

Setelah B1/B4 stabil, tambahkan B2 compact encoder sebagai semantic prior. Model kecil tidak menjadi core truth; embedding hanya memberi similarity prior yang dapat dibatalkan oleh konflik evidence. Uji FP32 dan quantized variants secara terpisah.

Adaptive refinement baru masuk setelah ada evidence bahwa ambiguity dapat dideteksi. Scheduler memilih crop dengan:

```text
value(crop) = expected_uncertainty_reduction(crop)
             / estimated_compute_cost(crop)
```

**Selesai jika:** B4/B5 mengurangi unsupported claims dan/atau memperbaiki risk-coverage tanpa melampaui budget CPU.

## Fase 6 — Benchmark, CPU tuning, dan publikasi

Run wajib dibandingkan pada hardware nyata. Catat cold start, warm p50/p95, RSS, thread scaling, image size, precision, model size, dan dataset split. Optimasi yang boleh dicoba setelah kualitas terukur meliputi cache, tiling, adaptive resolution, quantization, SIMD, dan parallel stages.

Publikasi dibangun dari kontribusi yang sempit dan bisa diuji:

```text
Evidence-grounded CPU visual reasoning
Provenance-aware scene graph with abstention
Adaptive region refinement under CPU budget
```

Paket publikasi harus memuat paper/report, code, benchmark manifest, configs, checksums, failure gallery, dan reproducer. Klaim “memahami semua gambar” cukup menjadi visi produk, bukan klaim ilmiah tanpa batas.

## Urutan commit yang disarankan

| Commit | Isi |
|---|---|
| `research: preserve legacy as B0` | Dokumentasi baseline dan branch policy |
| `feat(schema): add provenance scene graph contract` | Struct, JSON schema, golden tests |
| `feat(eval): add reproducible run manifest` | Evaluator skeleton dan run artifacts |
| `feat(imageview): add canonical image views` | Luminance/chroma/pyramid/coordinates |
| `feat(evidence): add low-level evidence fields` | Gradient/texture/contrast/edge |
| `feat(region): add graph region proposals` | Region formation dan merge ablation |
| `feat(graph): add visible geometric relations` | Scene graph dan relation tests |
| `feat(reasoning): add query planner and abstention` | Query operations dan calibration |
| `feat(semantic): add optional compact prior` | Model adapter, quantization experiments |
| `perf(runtime): add budget-aware refinement` | Crop scheduler dan CPU profiling |

## Langkah eksekusi pertama

Langkah pertama yang aman adalah mengimplementasikan **schema provenance + evaluator skeleton**, bukan detector baru. Setelah itu buat satu adapter B0 yang membungkus output HeraVision lama, sehingga evaluator langsung memiliki pembanding. Baru kemudian implementasikan B1 classical-only dan ukur apakah region/evidence lebih stabil daripada box UI lama.

Jika B1 belum dapat mengalahkan B0 pada track yang relevan atau belum dapat menjelaskan kegagalannya, kita tidak menambah semantic model. Dengan urutan ini, proyek berkembang berdasarkan bukti, bukan berdasarkan banyaknya kode atau besarnya klaim.
