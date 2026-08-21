# HeraVision Next — Dataset Protocol v0.1

## Dataset pilot

Pilot pertama menggunakan subset deterministik dari **COCO 2017 validation**. COCO dipilih karena menyediakan bounding boxes, object categories, segmentation, dan captioning pada gambar nyata; subset ini dipakai sebagai ground-truth geometry benchmark, bukan sebagai bukti bahwa B1 sudah memahami semua semantic category.

Sumber resmi:

- Dataset homepage: https://cocodataset.org/
- Validation images: https://images.cocodataset.org/zips/val2017.zip
- Instance annotations: https://images.cocodataset.org/annotations/annotations_trainval2017.zip
- COCO license/terms: https://cocodataset.org/#termsofuse

Data tidak disimpan di Git repository. Repository hanya menyimpan downloader, manifest, checksum, dan konfigurasi agar eksperimen dapat direproduksi tanpa membundel dataset besar.

## Sampling

Untuk pilot `N` gambar, daftar image COCO diurutkan berdasarkan `image_id`, kemudian dipilih dengan posisi yang tersebar merata dari seluruh validation set. Pemilihan tidak memakai prediction, sehingga tidak ada selection bias dari hasil engine. `N=300` digunakan sebagai pilot awal.

Setiap sample memiliki:

| Field | Makna |
|---|---|
| `id` | ID stabil berdasarkan COCO image ID |
| `image_path` | Path lokal relatif ke dataset root |
| `annotation_path` | File annotation resmi |
| `split` | `pilot` untuk eksperimen eksplorasi |
| `tags` | `coco`, `photo`, `bbox-groundtruth` |
| `sha256` | Digest image setelah download |
| `width`, `height` | Dimensi yang dikonfirmasi dari file |
| `source_url` | URL image resmi |

## Verifikasi

Sebelum run evaluator:

1. Semua image path harus ada.
2. SHA-256 lokal harus sama dengan nilai manifest.
3. File harus dapat didecode.
4. Width/height dari file harus sama dengan manifest.
5. Annotation JSON harus dapat dibaca dan memiliki image ID terkait.
6. Tidak boleh ada duplicate sample ID atau duplicate image hash.

## Apa yang diukur

Pada tahap B1 shape-neutral, metric utama adalah geometry: coverage terhadap ground-truth boxes, best IoU, boundary/center distance, fragmentation, dan false region rate. Semantic class accuracy belum dijadikan metric utama B1 karena B1 belum memiliki semantic classifier.

B0 dan B1 tetap dijalankan pada manifest dan input image yang sama. B0 dibandingkan sebagai legacy detector; B1 dibandingkan sebagai region/evidence engine. Hasil per-domain atau per-sample wajib disimpan bersama commit SHA, konfigurasi, hardware, dan runtime.

## Batas klaim

COCO validation subset bukan representasi seluruh dunia. Hasil pilot hanya boleh disebut **COCO pilot geometry result**. Klaim general vision membutuhkan domain tambahan, blind test, human review, dan failure analysis lintas foto, dokumen, diagram, screenshot, ilustrasi, dan gambar ambigu.

## Fallback pilot ringan: COCO128

Arsip pilot ringan diambil dari asset resmi Ultralytics: https://docs.ultralytics.com/datasets/detect/coco128 dan https://ultralytics.com/assets/coco128.zip. Dokumentasi menyebut COCO128 sebagai subset 128 image COCO train 2017 dengan anotasi object detection; arsip yang diterima berisi `images/train2017/*.jpg`, `labels/train2017/*.txt`, dan `LICENSE`. Format label adalah YOLO normalized bounding box: `class_id center_x center_y width height`.

COCO128 memiliki 128 gambar—cukup untuk pilot awal “ratusan” secara minimal, tetapi belum cukup untuk klaim generalisasi. Untuk evaluasi yang lebih kuat, COCO128 harus dilengkapi dataset lintas domain dan blind test yang lebih besar. Arsip COCO validation penuh tetap tidak dipakai karena host resmi mengalami `ERR_CERT_COMMON_NAME_INVALID` dan transfer sandbox sangat lambat; fallback ringan ini dicatat transparan, bukan disamarkan sebagai COCO validation penuh.
