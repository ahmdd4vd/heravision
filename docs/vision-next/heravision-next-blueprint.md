# HeraVision Next — Blueprint Riset Vision Engine CPU-First

**Status:** Perencanaan riset; belum mengubah source code utama  
**Basis:** Repository `ahmdd4vd/heravision`, ditambah kajian metode dan benchmark eksternal  
**Penulis:** Manus AI

## 1. Keputusan strategis

Kita akan melakukan **pivot total**. Prioritas “binary kecil” dihapus. HeraVision tidak lagi didefinisikan sebagai UI structure extractor, melainkan sebagai **evidence-first visual reasoning engine** yang mencoba memahami foto, dokumen, diagram, screenshot, ilustrasi, dan gambar ambigu melalui representasi visual berlapis.

Namun ada satu koreksi penting terhadap ambisi awal: “100000× lebih canggih dari manapun” bukan klaim yang dapat dipakai dalam riset sebelum ada baseline, dataset, hardware, metrik, dan eksperimen yang bisa mengalahkan pembanding. Target yang benar adalah menemukan **Pareto improvement yang dapat direproduksi**: kualitas grounding/reasoning lebih baik pada compute CPU tertentu, dengan uncertainty lebih jujur dan bukti yang dapat diaudit.

> Mesin tidak boleh berkata “saya melihat X” jika tidak dapat menunjukkan region dan evidence yang mendukung X.

## 2. Pergeseran pendekatan

HeraVision lama terutama melakukan `image → box → label`. Arsitektur baru menggunakan:

```text
image
  → canonical views
  → evidence fields
  → regions dan hypotheses
  → geometric relations
  → scene graph dengan provenance
  → query planning + rules + retrieval
  → grounded answer atau abstention
```

Dengan desain ini, kita tidak mencoba menyalin model vision besar secara langsung. Kita memecah masalah menjadi lima kemampuan yang lebih terukur: **mengukur**, **mengelompokkan**, **menghubungkan**, **memberi prior**, dan **menjawab berdasarkan bukti**.

## 3. Arsitektur yang diusulkan

| Lapisan | Isi | Mengapa diperlukan |
|---|---|---|
| Canonical views | Luminance, log-chroma, Gaussian pyramid, gradient, DoG/LoG, texture descriptor, optional block-DCT | Memisahkan sinyal warna, bentuk, tekstur, dan skala dengan biaya CPU rendah |
| Evidence field | Edge, corner, orientation, chroma, texture, flatness, saliency, contrast | Menyimpan fakta piksel sebelum memberi makna |
| Region engine | Over-segmentation konservatif lalu graph-based region merging | Menghasilkan region stabil tanpa menganggap setiap edge sebagai objek |
| Hypothesis engine | Kandidat objek dari contour, color, texture, symmetry, scale stability, optional encoder | Memelihara beberapa kemungkinan, bukan memaksa satu label |
| Relation engine | Posisi, containment, contact, overlap, alignment, occlusion, support | Mengubah region menjadi struktur adegan |
| Semantic prior | Mode `pure`, `compact`, dan `hybrid` | Menambahkan label/retrieval tanpa menjadikan neural model sumber kebenaran tunggal |
| Graph/query engine | Scene graph, rule engine, retrieval, query planner | Menjawab pertanyaan compositional secara murah dan dapat dilacak |
| Uncertainty | Entropy, cross-scale stability, evidence coverage, conflict rate | Memutuskan kapan menjawab dan kapan abstain |
| Refinement | Crop/zoom/feature tambahan berdasarkan information gain per compute | Mengarahkan CPU hanya ke region yang paling ambigu |

### Rumus inti awal

Untuk evidence vector pada lokasi `(x,y)`:

```text
E(x,y) = [edge, orientation, chroma, texture,
          flatness, saliency, saturation, local_contrast]
```

Untuk menggabungkan dua node region `i` dan `j`, gunakan cost yang bisa diaudit:

```text
w(i,j) = α·|Y_i - Y_j|
       + β·Δchroma(i,j)
       + γ·texture_distance(i,j)
       + δ·edge_barrier(i,j)
       + η·scale_inconsistency(i,j)
```

Untuk setiap hipotesis objek `h`:

```text
S(h | E) = b_type
         + Σ_k w_k · e_k(h)
         - λ_complexity · C(h)
         - λ_conflict · X(h)
```

Untuk confidence yang lebih jujur:

```text
Q = calibration(score)
    · stability_across_scale
    · stability_under_crop
    · evidence_coverage
    · (1 - conflict_rate)
```

Jika `Q` di bawah ambang jawaban, engine tidak mengarang. Ia mengembalikan `unknown`, region ambigu, konflik evidence, dan opsi refinement.

### Scene graph dengan provenance

```json
{
  "nodes": [
    {
      "id": "r17",
      "hypotheses": [
        {"label": "person", "score": 0.71},
        {"label": "statue", "score": 0.22}
      ],
      "region": {"bbox": [x, y, w, h], "mask_ref": "mask-17"},
      "evidence": ["boundary", "texture", "pose-like-symmetry"],
      "uncertainty": 0.29
    }
  ],
  "edges": [
    {
      "from": "r17",
      "to": "r22",
      "relation": "holding",
      "status": "inferred",
      "score": 0.58,
      "evidence": ["contact", "relative-position", "retrieved-prior"]
    }
  ]
}
```

Relasi `visible` dan `inferred` wajib dipisahkan. `holding` yang didukung contact geometry bukan hal yang sama dengan “benda itu kemungkinan milik orang tersebut” yang hanya berasal dari prior dunia.

## 4. Posisi model kecil dan matematika

Kita tidak akan menolak neural model secara dogmatis, tetapi menempatkannya di posisi yang tepat. MobileCLIP menunjukkan bahwa encoder image-text efisien dapat memberi trade-off latency-akurasi yang baik untuk zero-shot classification dan retrieval, tetapi angka latency/parameter dari paper harus diperlakukan sebagai hasil hardware dan benchmark tertentu, bukan jaminan untuk CPU kentang.[1]

MobileNetV4 menunjukkan bahwa efisiensi tidak cukup diukur dari MAC/FLOPs; pola akses memori dan operational intensity juga penting. Karena itu benchmark HeraVision harus mencatat bandwidth/memory behavior secara tidak langsung melalui peak RSS, thread scaling, p50/p95 latency, dan total end-to-end cost, bukan hanya waktu satu kernel.[2]

Quantization ONNX dapat menjadi jalur eksperimen untuk semantic prior. Dokumentasi resminya menjelaskan pemetaan 8-bit, format QOperator/QDQ, serta trade-off dynamic versus static quantization. Quantization tidak selalu mempercepat semua hardware dan dapat menimbulkan saturation atau penurunan akurasi, sehingga FP32, INT8 static, INT8 dynamic, dan INT4 harus diuji terpisah.[3]

Kesimpulan desainnya adalah **hybrid evidence-first**:

```text
classical math = observasi, boundary, region, geometry, stability
compact model = semantic prior, retrieval, open-vocabulary hints
symbolic graph = relations, query planning, provenance, abstention
```

## 5. Program eksperimen pertama

### Baseline

| Baseline | Isi |
|---|---|
| B0 | HeraVision lama |
| B1 | Classical-only: edge, region, color, texture, geometry |
| B2 | Compact encoder-only |
| B3 | Classical + compact semantic prior |
| B4 | B3 + scene graph/rules |
| B5 | B4 + adaptive crop/refinement |
| B6 | Model VLM lokal/eksternal lebih besar sebagai reference ceiling |

### Dataset suite

COCO menyediakan object detection, segmentation, recognition in context, captioning, dan keypoints, sehingga cocok untuk track region/object/caption dasar.[4] Visual Genome menyediakan objek, atribut, relasi, region descriptions, question answers, region graphs, dan scene graphs, sehingga cocok untuk track relational reasoning, tetapi anotasi crowd-sourced dan ketidakseimbangan predicate harus diperlakukan sebagai keterbatasan.[5]

Suite publik harus dilengkapi track dokumen/chart, diagram/screenshot, dan challenge set manusia untuk occlusion, low light, out-of-domain, misleading context, serta gambar yang memang tidak cukup informatif. Satu dataset tidak cukup untuk mendukung klaim “pemahaman umum”.

### Eksperimen prioritas

| ID | Pertanyaan | Hasil yang dicari |
|---|---|---|
| E1 | Apakah evidence stabil terhadap resize, brightness, compression, blur, dan crop? | Region/graph consistency lintas transformasi |
| E2 | Apakah fusion membantu dibanding classical-only dan encoder-only? | Akurasi naik tanpa hallucination/calibration memburuk |
| E3 | Apakah scene graph membantu pertanyaan compositional? | Relation/query accuracy dan provenance meningkat |
| E4 | Apakah adaptive refinement menghemat CPU? | Uncertainty turun dengan pixel/compute lebih sedikit |
| E5 | Apa efek FP32/INT8/INT4? | Pareto quality-latency-memory yang nyata |
| E6 | Apakah abstention dapat mendeteksi kegagalan? | Risk-coverage lebih baik dan unsupported claims turun |

### Metrik wajib

Kualitas region memakai IoU, boundary F-score, atau mask AP. Label memakai macro-F1 dan retrieval Recall@K. Relasi memakai precision/recall dan graph edit distance. Grounding memakai evidence IoU, provenance coverage, serta unsupported-claim rate. Uncertainty memakai ECE, Brier score, risk-coverage, dan AUROC error detection. Runtime memakai p50/p95 latency, cold start, peak RSS, thread scaling, dan quality per millisecond.

## 6. Aturan falsifikasi

Kita menganggap pendekatan gagal atau harus diubah arah jika perbaikan hanya muncul pada fixture buatan sendiri, accuracy naik tetapi unsupported-claim rate memburuk, “CPU-first” hanya tercapai di luar profil perangkat nyata, scene graph menghasilkan edge tanpa evidence region, adaptive refinement lebih mahal daripada compute yang dihemat, atau quantization tidak memberi manfaat end-to-end.

Aturan ini penting karena “temuan gila” yang tidak tahan falsifikasi hanya menjadi demo. Temuan yang layak dipublikasikan adalah temuan yang tetap muncul pada data publik, baseline berbeda, hardware berbeda, dan ablation yang menghapus komponen kunci.

## 7. Roadmap implementasi

| Fase | Deliverable | Kriteria selesai |
|---|---|---|
| R0 — Evaluator | Schema run, provenance JSON, hardware manifest, metric runner | Satu command menghasilkan prediction, evidence, metrics, latency |
| R1 — Perception core | Canonical views, evidence fields, region graph, geometry descriptors | B1 dapat diuji pada data nyata dan transformasi |
| R2 — Hypothesis engine | Multi-proposal object hypotheses dan cross-scale stability | Top-k hypothesis dengan evidence dan uncertainty |
| R3 — Scene graph | Node/edge schema, visible/inferred separation, graph queries | Subset QA compositional dapat dijawab dengan path evidence |
| R4 — Compact prior | Optional MobileCLIP/MobileNet-style encoder, quantized variants | B2/B3 dibandingkan fair dengan B1 |
| R5 — Adaptive runtime | Crop/refinement berbasis information gain | CPU quality-per-ms meningkat pada benchmark |
| R6 — Public release | Paper, benchmark card, code, failure gallery, reproducibility package | Orang lain dapat menjalankan dan memverifikasi hasil |

## 8. Strategi publikasi temuan

Publikasi harus berlangsung dalam beberapa lapisan, bukan menunggu satu klaim besar:

1. **GitHub research branch dan benchmark release.** Sertakan schema, command, dataset manifest, model checksum, failure gallery, dan hasil per commit.
2. **Technical report.** Tulis problem statement, metode, persamaan, baseline, ablation, hardware, dan keterbatasan. Jangan menulis “understands all images”; gunakan kontribusi yang spesifik.
3. **Preprint.** Setelah metode dan hasil cukup refereeable, arXiv dapat menjadi kanal distribusi preprint. Panduan resmi arXiv menekankan bahwa submission harus berupa kontribusi ilmiah yang topical dan refereeable, dilakukan oleh registered authors, mengikuti moderation, dan menerima lisensi distribusi.[6]
4. **Peer review.** OpenReview mendukung open peer review, open publishing, open access, discussion, dan API; venue serta deadline tetap harus diperiksa satu per satu.[7]
5. **Demo publik.** Buat visual debugger yang menampilkan gambar, region, graph, evidence, uncertainty, dan jawaban. Demo harus memperlihatkan failure, bukan hanya contoh terbaik.
6. **Replikasi.** Rilis `runs/<timestamp>-<git_sha>/` berisi `config.json`, `hardware.json`, `dataset_manifest.json`, `predictions.jsonl`, `evidence_graph.jsonl`, `metrics.json`, `latency.csv`, dan `failures.jsonl`.

Judul publikasi awal yang realistis misalnya **“Evidence-Grounded CPU Visual Reasoning with Adaptive Region Refinement”** atau **“A Provenance-Aware Neuro-Symbolic Scene Graph Engine for CPU-Constrained Visual Understanding.”** Judul seperti ini dapat diuji. Klaim “100000× lebih canggih dari manapun” tidak dapat diuji dan sebaiknya tidak dipakai sebagai judul ilmiah.

## 9. Struktur repository baru

```text
heravision/
  engine/              # runtime dan orchestration
  perception/          # canonical views, evidence fields
  regions/             # superpixels, graph merging, masks
  hypotheses/          # object proposals dan scoring
  relations/           # geometry, occlusion, support, alignment
  graph/               # scene graph + provenance schema
  semantic/            # optional compact encoder/retrieval
  reasoning/           # rules, query planner, abstention
  refinement/          # crop/zoom/information gain
  evaluator/           # datasets, metrics, calibration
  experiments/         # configs, ablations, reproducible runs
  paper/               # figures, tables, technical report
  demos/               # visual debugger dan examples
```

Repository lama sebaiknya dipertahankan sebagai `legacy-ui-detector` atau baseline B0. Jangan menghapusnya sebelum evaluator baru dapat membandingkan B0 dengan B1–B5 secara otomatis.

## 10. Keputusan langkah berikutnya

Kita **belum perlu menulis engine penuh sekarang**. Langkah teknis pertama yang paling bernilai adalah membuat evaluator dan schema provenance, kemudian mengimplementasikan B1 classical-only pada subset dataset nyata. Jika evaluator belum ada, kita tidak akan tahu apakah rumus baru benar-benar lebih baik atau hanya menghasilkan demo yang terlihat pintar.

Setelah B1 terukur, barulah kita menambahkan satu komponen pada satu waktu: semantic prior, graph reasoning, lalu adaptive refinement. Setiap penambahan harus mempunyai hipotesis, baseline, ablation, dan kondisi gagal.

### Referensi

[1]: https://machinelearning.apple.com/research/mobileclip — Apple Machine Learning Research, “MobileCLIP: Fast Image-Text Models through Multi-Modal Reinforced Training.”
[2]: https://arxiv.org/html/2404.10518v1 — Qin et al., “MobileNetV4 — Universal Models for the Mobile Ecosystem.”
[3]: https://onnxruntime.ai/docs/performance/model-optimizations/quantization.html — ONNX Runtime, “Quantize ONNX models.”
[4]: https://cocodataset.org/ — COCO official dataset site.
[5]: https://homes.cs.washington.edu/~ranjay/visualgenome/api.html — Visual Genome official dataset/API page.
[6]: https://info.arxiv.org/help/submit/index.html — arXiv, “Submission Guidelines.”
[7]: https://openreview.net/about — OpenReview, “About OpenReview.”
[8]: https://journals.sagepub.com/doi/abs/10.3233/NAI-240719 — Khan et al., “A survey of neurosymbolic visual reasoning with scene graphs and common sense knowledge.”
