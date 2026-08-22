# Candidate ambiguous visual observations

The first 20 ImageNet-style validation candidates inspected are mostly clear fishing photographs: fish, people holding fish, nets, and outdoor scenes. They are not genuinely ambiguous at the image-level. They may be useful as natural-photo hard negatives, but they should not be labeled `ambiguous` merely because the object classifier may be difficult.

This confirms that the local ImageNet-style validation source is a poor source for an honest ambiguous class in this selected slice. The fixture should not inflate ambiguous counts from these images. Genuine ambiguous data must be sourced separately, such as real blur, severe occlusion, tiny-object scenes, or mixed/low-quality images, and then reviewed explicitly.
