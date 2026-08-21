# HeraVision Next — Apa Langkah Berikutnya?

## Posisi kita sekarang

HeraVision Next sudah melewati **pilot geometry pertama**, bukan selesai sebagai vision engine umum. Pada 126 image COCO128, B0 legacy menghasilkan precision/F1 rendah tetapi stabil sebagai baseline, B1 raw menemukan lebih banyak region namun mengalami over-segmentation ekstrem, dan B1 filtered berhasil mengurangi false positive melalui logistic region filter kecil. Hasil ini membuktikan pipeline eksperimen bekerja; hasil ini belum membuktikan pemahaman semantic atau general vision.

Karena itu langkah berikutnya harus menjawab pertanyaan yang paling penting: **apakah B1 benar-benar memahami struktur visual, atau hanya menghasilkan region geometry yang kebetulan cocok pada subset COCO?**

## Fase berikutnya: blind multi-domain benchmark

Kita akan membuat dataset evaluasi baru yang tidak hanya berisi foto COCO. Dataset tersebut harus memiliki minimal lima domain:

| Domain | Tujuan pengujian |
|---|---|
| Foto natural | Object/region umum, occlusion, scale berbeda |
| Dokumen | Text blocks, table, form, chart, whitespace |
| Diagram | Node, connector, containment, alignment |
| Screenshot/UI | Mengetahui apakah engine tidak kembali menjadi UI-only detector |
| Low-quality/ambigu | Blur, compression, low light, crop, partial visibility |

Split-nya harus dipisah menjadi **development**, **calibration**, dan **blind test**. Threshold region filter tidak boleh diubah berdasarkan blind test. Setiap gambar harus memiliki source, hash, domain tag, dan annotation level yang jelas.

## Fase 1 — Memperbaiki region proposal

Pilot menunjukkan kelemahan paling besar bukan pada loading model, tetapi pada jumlah proposal B1 raw yang terlalu banyak. B1 raw menghasilkan sekitar 202 region per gambar, sedangkan B0 sekitar 18 region. Ini membuat recall naik sedikit tetapi precision runtuh.

Perbaikan yang akan diuji satu per satu:

1. **Boundary-aware merge.** Dua region tidak cukup dibandingkan berdasarkan warna/luminance; edge barrier dan stabilitas lintas scale harus masuk ke merge cost.
2. **Minimum meaningful area.** `MinArea=4` adalah pengaman awal, bukan solusi final. Nilai ini harus dikalibrasi berdasarkan resolusi dan domain.
3. **Multi-scale stability.** Region yang hanya muncul pada satu scale seharusnya mendapat uncertainty lebih tinggi atau dibuang.
4. **Region budget adaptif.** `MaxRegions=256` mencegah OOM, tetapi target berikutnya adalah memilih region berdasarkan information value, bukan memotong berdasarkan jumlah.
5. **Spatial relation pruning.** Relation builder tidak boleh membandingkan semua pasangan secara kuadratik jika region masih banyak. Gunakan spatial grid atau sweep-line candidate generation.

Setiap perubahan harus diuji sebagai `B1-old` versus `B1-new` pada test set yang sama. Jika precision membaik tetapi recall runtuh, perubahan tidak otomatis diterima.

## Fase 2 — Evidence dan provenance yang lebih kuat

Saat ini B1 dapat menyimpan evidence seperti region membership, geometry, boundary strength, dan learned filter score. Berikutnya setiap region perlu memiliki provenance yang lebih informatif:

```text
region
  → source pixels / polygon
  → scale observations
  → feature values
  → merge decisions
  → hypothesis score
  → relation evidence
```

Tujuannya adalah agar setiap jawaban dapat diaudit. Jika engine berkata dua region `above`, kita harus dapat menunjuk bounding geometry-nya. Jika engine berkata sebuah region `object`, harus ada semantic prior dan evidence yang jelas; jika belum ada, engine harus abstain atau memakai label netral.

## Fase 3 — Failure gallery dan human verification

Metric numerik saja belum cukup. Dari setiap run kita akan menyimpan:

| Artefak | Isi |
|---|---|
| `failures.jsonl` | Sample yang error atau metric-nya buruk |
| Contact sheet | Input, B0 boxes, B1 raw regions, B1 filtered regions, ground truth |
| Error class | Merge, fragmentation, missed region, wrong relation, decode/runtime |
| Human note | Apakah failure benar-benar salah atau annotation tidak lengkap |

Human verification dipakai terutama untuk novel region B1. Region baru tidak boleh dianggap discovery sebelum diperiksa terhadap ground truth atau reviewer.

## Fase 4 — Training berikutnya, jika data membenarkan

Training region filter sudah menunjukkan bahwa komponen kecil dapat membantu precision. Training berikutnya tidak boleh langsung menjadi VLM. Urutannya:

| Kandidat | Kapan dilatih | Target |
|---|---|---|
| Boundary merge scorer | Jika banyak fragmentation/merge error | Probabilitas dua region harus menyatu |
| Relation scorer | Jika geometry visible sudah stabil | Ranking `above`, `left_of`, `inside`, `touching` |
| Uncertainty calibrator | Jika score tidak sesuai risiko | Risk-coverage dan abstention |
| Compact semantic prior | Jika kebutuhan label semantic terbukti | Similarity prior, bukan sumber kebenaran tunggal |

Data training harus berasal dari development split. Calibration split digunakan untuk threshold. Blind test hanya dipakai sekali pada milestone evaluasi.

## Fase 5 — CPU dan stabilitas

Setelah kualitas region membaik, kita optimalkan CPU. Profiling harus memisahkan decode, canonical view, evidence, region merge, relation, learned filter, dan serialization. Metric minimal adalah cold start, warm p50/p95, peak RSS, pixel count, region count, dan thread count.

OOM pada `max-side=512` adalah temuan penting. Kita tidak akan menyelesaikannya hanya dengan menaikkan memory. Solusi arsitektural yang benar adalah membatasi intermediate graph, memakai spatial pruning, streaming feature field, dan adaptive resolution.

## Apa yang tidak dilakukan sekarang

Kita belum akan:

- Mengklaim HeraVision sudah memahami semua gambar.
- Mengganti B1 dengan model besar tanpa failure evidence.
- Melatih captioner atau VLM dari 126 image.
- Menyetel threshold berdasarkan blind test.
- Menganggap jumlah region yang lebih banyak sebagai kecanggihan.
- Menggabungkan branch ke `main` sebelum benchmark multi-domain stabil.

## Milestone sukses berikutnya

Milestone berikutnya disebut **B1-MD0**, bukan “AGI vision”. B1-MD0 dianggap selesai jika:

1. Setidaknya 300 image lintas domain memiliki manifest dan hash yang valid.
2. B0, B1 raw, dan B1 filtered selesai tanpa error atau semua error tercatat.
3. Ada blind test terpisah yang tidak dipakai untuk threshold tuning.
4. Precision/F1 B1 filtered tidak hanya membaik pada COCO128, tetapi juga pada minimal tiga domain tambahan.
5. Region count dan memory tidak meledak pada resolusi target.
6. Semua peningkatan memiliki ablation dan failure gallery.

Setelah milestone tersebut, kita baru dapat memilih apakah fokus berikutnya adalah boundary model, relation reasoning, semantic prior, atau perubahan fundamental pada representasi visual.
