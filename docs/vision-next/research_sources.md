# Temuan riset awal: CPU-first visual engine

## MobileCLIP — Apple, April 2024
Source: https://machinelearning.apple.com/research/mobileclip

Apple menjelaskan MobileCLIP sebagai keluarga model image-text efisien untuk runtime, dengan tujuan memperbaiki trade-off latency-akurasi zero-shot classification dan retrieval pada perangkat mobile. Halaman Apple menyatakan MobileCLIP-S2 2.3x lebih cepat dan lebih akurat dibanding model CLIP terbaik sebelumnya berbasis ViT-B/16 pada evaluasi yang mereka laporkan. Halaman terkait MobileCLIP2 menyebut keluarga MobileCLIP berada pada kisaran 3–15 ms latency dan 50–150M parameter, tetapi angka tersebut bergantung hardware, implementasi, dan benchmark; tidak boleh dipindahkan mentah-mentah ke CPU kentang.

Implikasi untuk HeraVision: gunakan encoder kecil semacam ini sebagai eksperimen baseline/teacher atau semantic embedding opsional, bukan sebagai satu-satunya “otak”. Core engine tetap perlu evidence geometry, region, uncertainty, dan reasoning yang bisa diaudit. Model image-text kecil dapat membantu open-vocabulary retrieval/classification, tetapi tidak otomatis memberi scene understanding lengkap.

## ONNX Runtime quantization
Source: https://onnxruntime.ai/docs/performance/model-optimizations/quantization.html

Dokumentasi ONNX Runtime mendefinisikan quantization 8-bit linear dengan pemetaan `val_fp32 = scale * (val_quantized - zero_point)`. Format yang tersedia meliputi QOperator dan QDQ; dynamic quantization menghitung parameter saat inference dan cenderung lebih fleksibel tetapi menambah overhead, sedangkan static quantization memakai calibration data dan dapat lebih efisien. Dokumentasi juga menekankan bahwa performa tidak selalu meningkat: hasil tergantung hardware, instruksi CPU, model, dan potensi saturation. S8S8 menjadi default, sementara U8U8 dapat dicoba bila saturation menurunkan akurasi pada arsitektur tertentu.

Implikasi untuk HeraVision: eksperimen CPU harus membandingkan FP32, INT8 static, INT8 dynamic, dan bila relevan INT4, dengan accuracy/latency/memory per perangkat. Quantization adalah optimasi runtime, bukan pendekatan reasoning baru; ia harus diletakkan di bawah representasi visual berlapis.

## Prinsip yang disimpan

1. Tidak ada klaim “CPU cepat” tanpa hardware dan benchmark yang disebutkan.
2. Encoder kecil cocok untuk semantic prior, retrieval, dan open-vocabulary baseline, tetapi harus digabungkan dengan symbolic evidence.
3. Quantization harus dievaluasi per operator/model/hardware; tidak boleh diasumsikan otomatis lebih cepat.
4. Setiap hasil eksperimen perlu menyimpan model, precision, calibration set, runtime, thread count, latency, memory, dan accuracy.

## MobileNetV4
Source: https://arxiv.org/html/2404.10518v1

Paper ini memperkenalkan Universal Inverted Bottleneck (UIB), yang menyatukan beberapa micro-architecture convolutional dan memberi fleksibilitas spatial/channel mixing, receptive field, serta compute trade-off. Paper juga memperkenalkan Mobile MQA yang dioptimalkan untuk pola akses memori accelerator, bukan hanya jumlah MAC, dan melaporkan speedup 39% pada konteks hardware mobile mereka. Model MNv4 dirancang melalui hardware-independent Pareto efficiency dan distillation.

Implikasi untuk HeraVision: pengukuran latency harus mencatat memory access dan bandwidth, bukan hanya FLOPs. Backbone ringan dapat dipakai sebagai semantic feature extractor atau teacher untuk distillation, sementara inti riset tetap berada pada evidence fusion, graph/geometry, retrieval, dan uncertainty.

## Neuro-symbolic visual reasoning with scene graphs
Source: https://journals.sagepub.com/doi/abs/10.3233/NAI-240719

Survey 2025 ini menempatkan scene graph generation sebagai representasi simbolik yang berisi object, visual relationship, dan attribute, lalu dapat menjadi basis visual question answering, image captioning, retrieval, dan event processing. Survey juga menekankan bahwa common-sense knowledge/knowledge graph dapat meningkatkan expressiveness dan reasoning, tetapi kualitas anotasi, predicate imbalance, variasi visual relationship, dan generalisasi ke scene baru tetap menjadi tantangan.

Implikasi untuk HeraVision: jangan membuat satu label “makna gambar” yang tidak dapat diaudit. Bangun scene graph dengan provenance: setiap node/edge menyimpan region evidence, feature yang mendukung, confidence, dan sumber prior. Pisahkan relasi yang terlihat langsung dari relasi yang hanya diinferensikan oleh common-sense rule, lalu izinkan abstention bila dukungan visual lemah.

## Dataset benchmark

### COCO
Source: https://cocodataset.org/

COCO menyediakan object detection, segmentation, recognition in context, stuff segmentation, captioning, dan human keypoints. Situs resminya menyebut 330K images, lebih dari 200K berlabel, 1.5M object instances, 80 object categories, 91 stuff categories, dan 5 captions per image. COCO cocok untuk baseline objek/region/caption, tetapi tidak cukup sebagai satu-satunya benchmark untuk “semua gambar”.

### Visual Genome
Source: https://homes.cs.washington.edu/~ranjay/visualgenome/api.html

Visual Genome menyediakan objek, atribut, hubungan, region descriptions, question answers, region graphs, dan scene graphs. Resource resmi menunjukkan skala file yang besar dan versi dataset yang berbeda. Ia cocok untuk supervised/evaluation target scene graph dan relasi, tetapi anotasi crowd-sourced dan predicate imbalance harus diperlakukan sebagai keterbatasan, bukan ground truth sempurna.

Implikasi: benchmark HeraVision harus berbentuk suite berlapis—COCO untuk objek/region/caption, Visual Genome untuk relasi/scene graph, dataset dokumen/diagram/screenshot untuk domain non-foto, dan challenge set buatan manusia untuk ambiguity/abstention. Jangan mengukur “pemahaman umum” hanya dari satu dataset.

## Publikasi

### arXiv
Source: https://info.arxiv.org/help/submit/index.html

Panduan resmi arXiv menyatakan submission harus berupa kontribusi ilmiah yang topical dan refereeable, dilakukan oleh registered authors, melewati moderation, dan membutuhkan persetujuan lisensi distribusi. Ini cocok sebagai kanal preprint setelah eksperimen memiliki metode, baseline, dan reproducibility package yang cukup.

### OpenReview
Source: https://openreview.net/about

OpenReview memosisikan diri sebagai platform open peer review, open publishing, open access, open discussion, dan API. Ia cocok untuk mengarahkan paper ke venue yang memakai proses review terbuka, tetapi venue, deadline, format, dan kebijakan review tetap harus dicek per call for papers.

Implikasi: publikasi HeraVision sebaiknya berlapis—GitHub release dan benchmark card untuk artefak, technical report/preprint untuk metode, blog/demo untuk akses publik, lalu peer-reviewed submission untuk klaim yang sudah kuat. Jangan mempublikasikan klaim “100000x” tanpa baseline; gunakan judul kontribusi yang spesifik seperti evidence-grounded CPU visual reasoning, adaptive region refinement, atau calibrated neuro-symbolic scene graph.
