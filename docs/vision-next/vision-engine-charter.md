# HeraVision Next: Research Charter

## Arah proyek

HeraVision generasi berikutnya tidak lagi diposisikan sebagai pendeteksi UI. Ia akan menjadi **mesin persepsi visual umum CPU-first**: menerima gambar, membangun representasi berlapis tentang isi, geometri, material, kejadian, dan ketidakpastian, lalu menjawab pertanyaan melalui inferensi yang dapat diaudit.

Target ini harus dibangun bertahap. Klaim “memahami semua gambar” dan “100000x lebih canggih” belum merupakan spesifikasi teknis; keduanya harus diubah menjadi benchmark, batas domain, dan metrik yang bisa mengalahkan baseline tertentu pada perangkat tertentu.

## Definisi masalah

Satu gambar tidak selalu cukup untuk menentukan kebenaran dunia. Mesin harus membedakan antara hal yang terlihat langsung, hal yang paling mungkin, dan hal yang tidak dapat disimpulkan. Karena itu output inti bukan sekadar label, melainkan:

| Lapisan | Pertanyaan yang dijawab | Tingkat kepastian |
|---|---|---|
| Observasi piksel | Warna, tepi, tekstur, kedalaman relatif, bentuk, saliency | Tinggi bila sinyal jelas |
| Entitas | Objek/permukaan/region apa yang konsisten secara visual | Sedang–tinggi |
| Relasi | Di mana objek berada, menyentuh, menutupi, menghadap, atau terhubung | Sedang |
| Kejadian/statis | Apa yang sedang terjadi atau bagaimana adegan tersusun | Sedang–rendah |
| Semantik | Apa makna, fungsi, atau konteks gambar | Bergantung domain |
| Ketidakpastian | Bagian mana yang ambigu atau tidak cukup didukung | Wajib selalu ada |

## Kendala desain

Istilah “CPU kentang” akan diwujudkan sebagai profil perangkat yang eksplisit, bukan slogan. Profil awal yang perlu dipakai untuk eksperimen adalah CPU x86-64 empat thread, RAM 8 GB, tanpa GPU, dan target latensi yang dilaporkan terpisah untuk gambar kecil, sedang, dan besar. Semua eksperimen wajib mencatat waktu wall-clock, peak memory, jumlah thread, ukuran model atau tabel, serta kualitas output.

Arsitektur boleh menggunakan matematika klasik, lookup table, kompresi, cache, dan model kecil yang berjalan di CPU. Namun kita tidak boleh berpura-pura bahwa algoritma deterministik murni dapat menggantikan seluruh pengetahuan visual dunia tanpa trade-off. Mesin harus memiliki mode “tidak tahu”, abstention, dan daftar bukti yang mendukung setiap kesimpulan.

## Sasaran bertahap

| Tahap | Sasaran | Bukti keberhasilan |
|---|---|---|
| P0 | Persepsi umum tingkat rendah | Segmentasi region, kontur, tekstur, warna, dan invariant geometri stabil pada transformasi dasar |
| P1 | Scene graph visual | Entitas dan relasi dapat direkonstruksi pada dataset terkontrol, bukan hanya screenshot UI |
| P2 | Deskripsi grounded | Caption/fakta menunjuk region bukti dan tidak mengarang objek yang tidak terlihat |
| P3 | QA visual CPU | Pertanyaan faktual tentang gambar dijawab dengan akurasi dan latensi yang terukur |
| P4 | Generalisasi | Kinerja tetap terbaca pada foto, ilustrasi, dokumen, diagram, screenshot, dan gambar ambigu |
| P5 | Temuan baru | Ada metode atau kombinasi metode yang menunjukkan Pareto improvement atas baseline terbuka pada CPU dan benchmark yang sama |

## Prinsip anti-hype

Tidak ada klaim “lebih canggih” tanpa baseline, dataset, seed, perangkat, metrik, dan artefak reproduksi. Jika sebuah pendekatan hanya bekerja pada fixture buatan sendiri, hasilnya disebut prototipe atau overfit, bukan terobosan. Publikasi temuan harus menyertakan kegagalan, contoh kontra, dan kondisi ketika engine memilih abstain.

## Pertanyaan riset inti

1. Apakah representasi visual bertingkat yang menggabungkan kontur, region, tekstur, geometri, dan memori dapat memberi pemahaman scene yang lebih baik per watt CPU daripada satu pipeline deteksi langsung?
2. Dapatkah objek dan relasi ditemukan melalui konsistensi lintas skala, transformasi, dan crop tanpa model besar?
3. Bagaimana cara mengubah evidence map menjadi jawaban bahasa yang grounded, sehingga setiap klaim dapat dilacak kembali ke pixel/region?
4. Seberapa jauh dictionary visual, program induksi, dan retrieval contoh dapat menggantikan parameter model besar pada domain tertentu?
5. Bagaimana mesin mengukur ketidakpastian dan menolak menjawab ketika gambar memang tidak cukup informatif?

## Definisi “engine sendiri”

“Engine sendiri” berarti arsitektur, representasi, algoritma penggabungan evidence, evaluator, dan runtime dimiliki serta dapat diaudit oleh proyek. Itu tidak mengharuskan setiap komponen ditulis tanpa teori atau tanpa model kecil. Bila sebuah eksperimen menggunakan model pretrained atau dataset eksternal, ketergantungan tersebut harus dinyatakan terang dan dibandingkan dengan baseline tanpa komponen itu.

## Output riset pertama

Eksperimen awal harus menghasilkan tiga artefak: (1) spesifikasi representasi internal dan kontrak JSON, (2) benchmark CPU dengan dataset lintas domain dan baseline, serta (3) prototipe kecil yang dapat menunjukkan satu peningkatan terukur. Fokus pertama bukan membuat chatbot vision lengkap, melainkan membuktikan bahwa **evidence berlapis + reasoning geometris + abstention** lebih berguna daripada detector UI lama untuk domain gambar yang lebih luas.
