import sys

from humanloop.data.lerobot import export as lerobot_export
from humanloop.data.rlds import export as rlds_export
from humanloop.data.parquet import convert as parquet_convert

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("usage: export.py lerobot <out_dir> [task_description] <hmdf_file> [hmdf_file ...]")
        print("       export.py rlds    <out_dir> <hmdf_file> [hmdf_file ...]")
        print("       export.py parquet <input_dir> <output.parquet>")
        sys.exit(1)

    fmt = sys.argv[1]

    if fmt == "lerobot":
        out_dir = sys.argv[2]
        task_desc = sys.argv[3] if len(sys.argv) > 4 else ""
        paths = sys.argv[4:] if len(sys.argv) > 4 else [sys.argv[3]]
        lerobot_export(paths, out_dir, task_desc)

    elif fmt == "rlds":
        out_dir = sys.argv[2]
        paths = sys.argv[3:]
        rlds_export(paths, out_dir)

    elif fmt == "parquet":
        parquet_convert(sys.argv[2], sys.argv[3])
