# MD1 Annotation Review Notes

## Imagewoof

Sepuluh fixture Imagewoof didominasi satu anjing dengan variasi crop, pose, occlusion, background, dan interaksi manusia. Region utama dapat dianotasi sebagai visible foreground dog; bagian tubuh yang terpotong atau tertutup perlu memakai visibility `partial` atau `uncertain` bila batas box tidak jelas.

## Wikimedia

Fixture Wikimedia mencakup flowchart, capacitor schematic, pie chart, blank scan, eye schematic, image-format infographic, network diagram, rocket-engine schematic, dan tracking flow chart. Diagram memiliki banyak komponen internal. Untuk tahap pertama, anotasi region harus membedakan `whole_diagram` dari subregion hanya jika subregion punya batas visual yang jelas; blank scan sebaiknya diberi `ignore` atau satu region background dengan alasan `no_visible_structure`.

## Konsistensi

Annotation schema harus menyimpan `visibility`, `confidence`, `review_status`, dan `source`. Tidak boleh memakai class folder Imagenette/Imagewoof sebagai semantic ground truth bbox. Untuk diagram, whole-canvas region dan node-level region harus dipisah melalui `region_type`.
