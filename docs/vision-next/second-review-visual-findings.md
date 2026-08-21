# Second Review Visual Findings — Pass 1

## Imagenette

Overlay menunjukkan sebagian besar box mencakup objek utama dengan batas yang masuk akal. Kota dan scene ambigu memakai box besar yang mencakup struktur utama, sehingga statusnya lebih tepat `needs-review` daripada accepted final. Beberapa objek seperti radio, fish, dan truck memiliki box yang cukup jelas.

## Imagewoof

Box anjing umumnya mengikuti foreground utama. Namun ada crop close-up, dua anjing/foreground overlap, dan kasus manusia+anjing yang membuat batas target tidak selalu independen. Sample tersebut harus tetap `needs-review` atau `partial`, bukan accepted tanpa catatan.

## Review decision

Pass kedua tidak boleh langsung mengubah semua box menjadi accepted. Sample dengan batas visual jelas dapat accepted provisional; scene/crop/occlusion ambiguous masuk needs-review; blank scan tetap ignore. Disagreement log harus menyimpan alasan dan tidak menghapus box provisional.
