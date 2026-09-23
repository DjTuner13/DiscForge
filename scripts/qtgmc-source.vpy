import os

import vapoursynth as vs
from vsdeinterlace import QTempGaussMC

source = os.environ.get("DISCFORGE_SOURCE")
if not source:
    raise RuntimeError("DISCFORGE_SOURCE is required")

core = vs.core
core.max_cache_size = 2048
src = core.lsmas.LWLibavSource(source=source)
out = QTempGaussMC().deinterlace(src, tff=True)
out.set_output()
