import os

from test_utils.test_utils import run_packager

from test_utils.common import PackagerReturnCode


def test_01_run_without_command(packager_binary):
    """Run packager without any command"""
    result = run_packager(
        packager_binary, "", expected_returncode=PackagerReturnCode.CMD_LINE_ERROR
    )
    stdout = result.communicate()[0]

    assert "ERROR" in stdout


def test_02_run_with_invalid_command(packager_binary):
    """Run packager with an invalid command"""
    result = run_packager(
        packager_binary,
        "invalid-command",
        expected_returncode=PackagerReturnCode.CMD_LINE_ERROR,
    )
    stdout = result.communicate()[0]

    assert "ERROR" in stdout


def test_03_run_with_nonexisting_image(packager_binary, context):
    """Run packager with a non-existing image"""

    result = run_packager(
        packager_binary,
        "build-image",
        context=context,
        image_name="nonexisting_image",
        expected_returncode=PackagerReturnCode.DEFAULT_ERROR,
    )
    stdout = result.communicate()[0]

    assert "ERROR" in stdout
    assert "Failed to build Docker image:" in stdout


def test_04_run_with_nonexisting_package(test_image, packager_binary, test_repo, context):
    """Run packager with a non-existing package"""

    result = run_packager(
        packager_binary,
        "build-package",
        context=context,
        image_name=test_image,
        name="nonexisting_package",
        output_dir=test_repo,
        expected_returncode=PackagerReturnCode.DEFAULT_ERROR,
    )
    stdout = result.communicate()[0]

    assert "ERROR" in stdout


def test_05_run_with_nonexisting_app(test_image, packager_binary, test_repo, context):
    """Run packager with a non-existing app"""

    result = run_packager(
        packager_binary,
        "build-app",
        context=context,
        image_name=test_image,
        name="nonexisting_app",
        output_dir=test_repo,
        expected_returncode=PackagerReturnCode.DEFAULT_ERROR,
    )
    stdout = result.communicate()[0]

    assert "ERROR" in stdout


def test_06_build_package_without_name_or_all(test_image, packager_binary, context, test_repo):
    """Run build-package without --name or --all"""

    result = run_packager(
        packager_binary,
        "build-package",
        context=context,
        image_name=test_image,
        output_dir=test_repo,
        expected_returncode=PackagerReturnCode.CMD_LINE_ERROR,
    )
    stdout = result.communicate()[0]

    assert "ERROR" in stdout


def test_07_build_app_without_name_or_all(test_image, packager_binary, context, test_repo):
    """Run build-app without --name or --all"""

    result = run_packager(
        packager_binary,
        "build-app",
        context=context,
        image_name=test_image,
        output_dir=test_repo,
        expected_returncode=PackagerReturnCode.CMD_LINE_ERROR,
    )
    stdout = result.communicate()[0]

    assert "ERROR" in stdout
