# type: ignore
# pylint: disable=no-value-for-parameter,import-outside-toplevel,import-error,no-member,missing-function-docstring

from kfp import dsl

from .consts import PYTHON_IMAGE, TOOLBOX_IMAGE


@dsl.container_component
def pvc_to_mt_bench_op(mt_bench_output: dsl.Output[dsl.Artifact], pvc_path: str):
    return dsl.ContainerSpec(
        TOOLBOX_IMAGE,
        ["/bin/sh", "-c"],
        [f"cp -r {pvc_path} {mt_bench_output.path}"],
    )


@dsl.container_component
def pvc_to_model_op(model: dsl.Output[dsl.Model], pvc_path: str):
    return dsl.ContainerSpec(
        TOOLBOX_IMAGE,
        ["/bin/sh", "-c"],
        [f"cp -r {pvc_path} {model.path}"],
    )


@dsl.component(
    base_image=PYTHON_IMAGE,
    install_kfp_package=False,
    packages_to_install=["huggingface_hub"],
)
def huggingface_importer_op(repo_name: str, model_path: str = "/model"):
    from huggingface_hub import snapshot_download

    snapshot_download(repo_id=repo_name, cache_dir="/tmp", local_dir=model_path)
