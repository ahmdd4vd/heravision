# Program Eksperimen HeraVision Next

## Tujuan eksperimen

Eksperimen pertama tidak bertujuan membuktikan bahwa HeraVision “memahami semua gambar”. Tujuannya adalah menguji hipotesis yang lebih tajam: **representasi evidence berlapis, graph reasoning, dan inference adaptif dapat memberi pemahaman visual yang lebih grounded per unit compute dibanding detektor UI tunggal atau encoder kecil yang dipakai sendirian.**

## Profil hardware wajib

Setiap hasil harus menyebutkan CPU model, jumlah core/thread, RAM, operating system, compiler/runtime, thread count, precision, ukuran input, dan warm/cold latency. Minimal jalankan satu profil “potato CPU” yang nyata, misalnya CPU x86-64 empat thread tanpa GPU dan RAM 8 GB. Jangan memakai angka dari hardware Apple, mobile accelerator, atau GPU sebagai klaim CPU HeraVision.

## Dataset dan split

| Track | Dataset/jenis data | Target |
|---|---|---|
| Region/object | COCO | Region, object, segmentation proxy, caption grounding |
| Relation/scene graph | Visual Genome | Object-attribute-relation graph |
| Visual reasoning | GQA atau subset berlisensi jelas | Query composition dan relation traversal |
| Dokumen | DocVQA/ChartQA atau subset resmi | Teks, chart, tabel, layout |
| Diagram/UI | Fixture nyata beragam, bukan hanya synthetic UI | Diagram topology, screenshot, form, code |
| Foto/ambiguity | Challenge set manusia | Occlusion, low light, unusual objects, misleading context |

COCO dipakai sebagai baseline objek/region/caption; Visual Genome untuk relasi dan graph; track dokumen serta diagram menjaga agar engine tidak kembali overfit ke foto atau UI. Split harus dipisahkan per gambar asli, bukan hanya per crop, agar tidak terjadi leakage.

## Baseline yang harus dibandingkan

| Baseline | Deskripsi | Tujuan |
|---|---|---|
| B0 | HeraVision lama | Menunjukkan batas detector UI saat dipakai di luar domain |
| B1 | Classical-only | Edge, region, color, texture, geometry, tanpa semantic encoder |
| B2 | Compact encoder-only | MobileCLIP/MobileNet-style embedding, retrieval/classification sederhana |
| B3 | Classical + compact encoder | Fusion semantic prior dengan evidence geometry |
| B4 | B3 + scene graph/rules | Mengukur keuntungan reasoning eksplisit |
| B5 | B4 + adaptive refinement | Mengukur keuntungan crop/zoom berbasis information gain |
| B6 | Model VLM eksternal atau model lokal lebih besar | Ceiling/reference, bukan target CPU utama |

Baseline harus memiliki preprocessing dan input budget yang setara. Bila model eksternal tidak dapat dijalankan pada profil CPU, laporkan sebagai reference offline dengan alasan dan jangan memasukkannya ke tabel latency yang sama.

## Eksperimen inti

### E1 — Apakah evidence lebih stabil daripada label langsung?

Berikan transformasi yang tidak mengubah isi: resize, brightness shift, contrast shift, JPEG compression, blur ringan, crop margin, dan color temperature. Ukur apakah region, boundary, dan relasi dasar tetap stabil. Metrik utama adalah mask/box consistency, graph edit distance, dan calibration.

**Hipotesis:** evidence geometry dan region lebih stabil terhadap sebagian transformasi dibanding label semantic langsung.  
**Falsifikasi:** jika B1 tidak lebih stabil daripada B2/B3 atau stabilitas hanya terjadi pada fixture synthetic.

### E2 — Apakah fusion membantu?

Bandingkan B1, B2, dan B3 pada object/region, open-vocabulary retrieval, dan scene graph. Gunakan ablation: hapus warna, texture, geometry, scale stability, dan semantic prior satu per satu.

**Hipotesis:** semantic prior membantu penamaan, tetapi evidence geometry mengurangi false positive dan hallucination.  
**Falsifikasi:** B3 tidak mengalahkan B1 dan B2 setelah compute budget dinormalisasi, atau fusion meningkatkan accuracy tetapi merusak calibration.

### E3 — Apakah scene graph membantu pertanyaan compositional?

Gunakan subset pertanyaan yang dapat diterjemahkan menjadi query graph. Bandingkan jawaban direct retrieval dengan graph traversal. Simpan path evidence untuk setiap jawaban.

**Hipotesis:** graph reasoning lebih unggul pada pertanyaan relasional sederhana dan lebih mudah diaudit.  
**Falsifikasi:** graph edit/answer accuracy tidak meningkat, atau engine menjawab dengan benar tetapi provenance kosong/tidak konsisten.

### E4 — Apakah adaptive refinement memberi Pareto improvement?

Ukur tiga mode: full-resolution fixed, crop semua region, dan crop hanya region dengan nilai information gain tertinggi. Catat jumlah pixel yang diproses, waktu, memory, accuracy, dan perubahan uncertainty.

**Hipotesis:** sebagian besar compute dapat dipindahkan ke region ambigu tanpa penurunan kualitas.  
**Falsifikasi:** refinement tidak mengurangi uncertainty, atau biaya planning melebihi compute yang dihemat.

### E5 — Precision dan runtime

Jalankan FP32, INT8 static, INT8 dynamic, dan INT4 jika runtime/operator mendukung. Gunakan calibration set yang tidak bocor ke test. Ukur accuracy, latency p50/p95, peak RSS, binary/model size, dan error rate.

**Hipotesis:** INT8 memberi Pareto improvement pada compact semantic prior, tetapi tidak otomatis pada seluruh operator atau hardware.  
**Falsifikasi:** quantization tidak mempercepat runtime atau menurunkan accuracy/calibration melewati threshold yang dapat diterima.

### E6 — Abstention dan kalibrasi

Campur gambar mudah, ambigu, out-of-domain, occluded, dan low-quality. Mesin harus dapat menjawab “unknown” dan menunjuk region yang tidak cukup bukti. Ukur selective accuracy, risk-coverage curve, ECE/Brier score, AUROC untuk error detection, serta kualitas provenance.

**Hipotesis:** evidence conflict dan cross-scale instability dapat memprediksi kegagalan lebih baik daripada raw classifier score.  
**Falsifikasi:** abstention tidak meningkatkan kualitas pada coverage yang sama, atau confidence tetap tinggi saat evidence conflict tinggi.

## Metrik utama

| Dimensi | Metrik |
|---|---|
| Region | IoU, boundary F-score, mask AP bila anotasi tersedia |
| Object/label | Accuracy, macro-F1, open-vocabulary retrieval Recall@K |
| Relation | Relation precision/recall, graph edit distance, triplet accuracy |
| Caption/QA | Exact/soft accuracy, grounded answer rate, human factuality |
| Grounding | Evidence IoU, provenance coverage, unsupported-claim rate |
| Uncertainty | ECE, Brier, risk-coverage, abstention utility |
| Runtime | p50/p95 latency, cold start, peak RSS, thread scaling |
| Efisiensi | quality per millisecond, quality per MB, pixels processed per answer |

Satu skor gabungan boleh dibuat untuk dashboard internal, tetapi leaderboard publik harus tetap memperlihatkan semua dimensi. Skor gabungan yang menyembunyikan hallucination atau latency tidak boleh menjadi klaim utama.

## Protokol reproduksi

Setiap run menghasilkan directory berikut:

```text
runs/<timestamp>-<git_sha>/
  config.json
  hardware.json
  dataset_manifest.json
  predictions.jsonl
  evidence_graph.jsonl
  metrics.json
  latency.csv
  failures.jsonl
  README.md
```

`README.md` run wajib menjawab: commit apa yang dijalankan, dependency apa yang dipakai, model/weights apa, calibration set apa, command exact, dan apakah hasil warm/cold. Semua random seed, preprocessing, dan threshold dicatat.

## Aturan falsifikasi

Riset harus berhenti atau berubah arah bila salah satu kondisi berikut terjadi:

1. Perbaikan hanya muncul pada fixture buatan sendiri dan hilang pada COCO/Visual Genome/challenge set.
2. Accuracy naik tetapi unsupported-claim rate atau calibration memburuk secara signifikan.
3. “CPU-first” hanya tercapai dengan mengorbankan latency yang tidak realistis atau memory yang melebihi profil perangkat.
4. Scene graph menghasilkan banyak edge yang tidak memiliki evidence region.
5. Adaptive refinement menambah kompleksitas lebih besar daripada manfaatnya.
6. Model kecil/quantization tidak memberi keuntungan setelah pengukuran end-to-end, bukan hanya kernel benchmark.

## Urutan eksperimen yang disarankan

1. Bangun evaluator dan schema provenance sebelum menambah algoritma.
2. Reimplementasikan B0 dan B1 pada repository baru sebagai baseline yang dapat diulang.
3. Buat B3 fusion dengan semantic prior opsional.
4. Tambahkan scene graph dan query planner untuk subset pertanyaan yang terdefinisi.
5. Jalankan E1–E3 pada data kecil namun nyata.
6. Baru lakukan refinement, quantization, dan optimasi runtime.
7. Publikasikan kegagalan dan ablation sebelum mengklaim kemampuan umum.
