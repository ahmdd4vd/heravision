# MD1 Visual Findings

## Failure gallery

Failure gallery pertama didominasi Imagewoof dog images dengan variasi pose, occlusion, pagar, manusia, dan background kompleks. Ini cocok untuk menguji fragmentation, foreground/background separation, serta region proposal yang terlalu banyak. Gallery menunjukkan sample nyata dan beragam, bukan fixture sintetis.

## Imagenette annotation subset

Subset Imagenette berisi objek dan konteks beragam: radio, kota, manusia, aktivitas luar ruang, kendaraan, mesin, ikan, dan objek rumah tangga. Sepuluh sample ini layak digunakan sebagai annotation fixture awal karena bounding region visible dapat ditinjau tanpa memaksakan label semantic baru.

## Decision

Annotation MD1 harus menggunakan region visible/foreground yang dapat ditelusuri, bukan menyalin class label sebagai bounding box. Failure gallery akan dipakai untuk memilih failure family, sementara subset 30 image tetap menjadi fixture terpisah dari blind 300 final.
