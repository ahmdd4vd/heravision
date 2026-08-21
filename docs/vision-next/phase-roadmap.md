# HeraVision Next — Phase Roadmap dan Exit Criteria

## Status objektif saat ini

Branch `vision-next` pada commit `b1fe10c` adalah **fondasi awal**, bukan engine vision selesai. Saat ini branch memiliki 17 source Go baru, 8 file test, dan hanya satu fixture repository (`testdata/ui.png`). Karena itu belum ada bukti bahwa B1 lebih canggih daripada B0, belum ada benchmark ratusan gambar, dan belum ada training model.

## Fase 0 — Baseline freeze

**Tujuan:** membuat titik pembanding yang tidak berubah.

| Item | Keputusan |
|---|---|
| B0 | Engine lama pada commit/tag yang dibekukan |
| B1 | Engine `internal/visionnext` pada commit eksperimen |
| Input | Manifest yang sama untuk kedua engine |
| Output | Observation/SceneGraph yang sama |
| Runtime | Proses/worker terpisah agar cache dan global state tidak bias |
| Status selesai | B0 dan B1 dapat dijalankan pada satu sample dan output JSON tersimpan |

Tidak boleh mengklaim peningkatan sebelum fase ini selesai.

## Fase 1 — Evaluator nyata

**Tujuan:** membangun pengukuran sebelum memperbanyak algoritma.

Deliverable-nya adalah `heravision eval` atau runner ekuivalen yang membaca manifest, menjalankan B0/B1, menyimpan prediction, graph, latency, failure, dan metrics per sample.

Exit criteria:

1. Evaluator dapat menjalankan minimal 10 gambar nyata.
2. Semua run memiliki `git_sha`, hardware, config, input dimensions, warm/cold timing, dan error log.
3. Matching geometry-only memakai IoU dan tidak bergantung pada label UI.
4. Ada golden output untuk satu sample dan regression test untuk schema.

## Fase 2 — Dataset pilot 100 gambar

**Tujuan:** mengetahui apakah B1 bekerja di luar satu fixture.

Komposisi pilot yang disarankan:

| Domain | Jumlah awal | Contoh |
|---|---:|---|
| Foto natural | 25 | Indoor, outdoor, manusia, benda, kendaraan |
| Dokumen | 20 | Halaman teks, tabel, formulir, chart |
| Diagram | 15 | Flowchart, architecture, mind map |
| Screenshot/UI | 15 | Web, mobile, terminal, editor |
| Ilustrasi/komik | 10 | Gambar non-fotorealistik |
| Low-quality/ambigu | 15 | Blur, occlusion, low light, compression |
| **Total** | **100** | Split eksplorasi, bukan klaim publik final |

Dataset pilot harus memiliki sumber/legalitas yang tercatat, hash file, manifest, domain tag, dan minimal annotation ringan untuk subset evaluasi. Jangan menganggap 100 gambar tanpa annotation sebagai benchmark kualitas.

Exit criteria:

1. Seluruh 100 gambar berhasil diproses atau masuk failure log yang dapat dijelaskan.
2. Tidak ada panic, deadlock, atau output JSON invalid.
3. Tersedia contact sheet/failure gallery untuk review manusia.
4. B0 dan B1 mempunyai tabel per-domain untuk region count, IoU match, novelty, fragmentation, relation count, latency, dan memory.

## Fase 3 — Dataset evaluasi 500 gambar

**Tujuan:** menguji generalisasi sebelum perubahan algoritma besar.

Komposisi harus memperluas pilot dan memisahkan train/calibration/test. Untuk dataset publik seperti COCO atau Visual Genome, simpan hanya manifest dan downloader resmi bila ukuran/licensing tidak memungkinkan menyimpan data di repository.

Target minimal:

| Split | Jumlah | Kegunaan |
|---|---:|---|
| Development | 250 | Debug dan tuning |
| Calibration | 100 | Threshold/confidence/abstention |
| Blind test | 150 | Tidak boleh dipakai untuk tuning |

Exit criteria:

1. B1 memiliki hasil terpisah untuk setiap domain.
2. Semua threshold yang diubah dicatat sebelum blind test.
3. Ada human review pada sample novel dan failure.
4. Tidak ada klaim global bila hanya satu domain yang membaik.

## Fase 4 — Perception improvement berbasis failure

**Tujuan:** memperbaiki region, hypothesis, dan relation berdasarkan failure nyata, bukan menambah fitur secara acak.

Loop pengembangan:

```text
run 500 images
  → cluster failures
  → pilih satu failure family
  → tulis hypothesis perubahan
  → implementasi satu perubahan
  → ablation B1-old vs B1-new
  → rerun fixed test set
  → keep/revert berdasarkan metrik
```

Failure family yang mungkin adalah region fragmentation, region merge, boundary leak, scale instability, false relation, atau low-quality collapse.

Exit criteria: setiap perubahan memiliki issue/experiment ID, before/after metrics, contoh visual, dan keputusan keep/revert.

## Fase 5 — Training hanya bila dibuktikan perlu

Training **tidak dimulai dari nol secara membabi buta**. Pertama kita bandingkan tiga opsi:

| Opsi | Dipakai untuk |
|---|---|
| No training | Geometry/evidence murni |
| Training komponen kecil | Region merge, boundary score, calibration, relation ranking |
| Compact pretrained model | Semantic prior/open-vocabulary retrieval |

Training diperbolehkan hanya jika failure analysis menunjukkan rule/matematika tidak cukup dan ada dataset/annotation yang legal serta cukup. Komponen pertama yang layak dilatih bukan VLM penuh, melainkan model kecil untuk satu tugas terdefinisi, misalnya:

```text
input: pair of adjacent regions + evidence features
output: merge_probability
```

atau:

```text
input: pair of region nodes + geometric features
output: relation score untuk visible predicate
```

Exit criteria:

1. Model kecil mengalahkan rule baseline pada blind test.
2. Generalisasi diuji lintas domain.
3. Model tidak meningkatkan unsupported claim rate.
4. Ukuran, latency, dan memory tetap sesuai CPU profile.

## Fase 6 — Benchmark CPU dan regresi

Setelah kualitas stabil, jalankan benchmark pada CPU target dengan warm/cold runs, p50/p95, peak RSS, thread scaling, dan input-size scaling. Semua commit baru wajib menjalankan unit test, integration test, evaluator smoke, dan regression set.

Exit criteria:

1. Tidak ada regression tanpa catatan.
2. Hasil B0/B1 dapat direproduksi dari commit dan manifest.
3. Ada ablation table untuk setiap komponen utama.
4. Ada failure gallery dan daftar keterbatasan.

## Fase 7 — Publikasi dan rilis

Hanya setelah fase 3–6 yang memadai, kita menulis technical report/preprint. Claim harus spesifik, misalnya “evidence-grounded region reasoning under CPU budget”, bukan “memahami semua gambar”. Artefak publik mencakup code, manifest, config, metrics, checksum, benchmark card, dan failure examples.

## Aturan status

| Status | Arti |
|---|---|
| `foundation` | Kontrak dan modul awal tersedia, belum ada bukti generalisasi |
| `pilot-tested` | Sudah diproses 100 gambar dan failure gallery tersedia |
| `benchmark-tested` | Sudah diuji 500+ gambar dengan blind test dan metrics |
| `trained-component` | Komponen kecil dilatih dan mengalahkan baseline pada blind test |
| `research-result` | Ada ablation, reproducibility package, dan klaim yang dapat dipertahankan |

**Status branch setelah pilot COCO128: `trained-component` untuk pilot geometry saja.** Ini belum berarti benchmark general vision selesai; domain tambahan dan blind test eksternal masih wajib.
