OUTPUT_DIR:=out
REPO_DIR:=`pwd`
E2E_RESULTS_DIR:=${OUTPUT_DIR}/e2e/results
E2E_RUNTIME_DIR:=${OUTPUT_DIR}/e2e/runtime
E2E_TESTS_DIR:=tests
E2E_PYTHON_VENV_DIR:=.venv
E2E_REQUIREMENTS_FILE:=e2e-requirements.txt

INCLUDETAG_F:=
ifdef TAG
	INCLUDETAG_F:=--include ${TAG}
endif

test: clean
	@mkdir -pv ${E2E_RESULTS_DIR} ${E2E_RUNTIME_DIR}
	@sh -c "export PATH=\"\$$PATH:\$$HOME/go/bin\" && . ${E2E_PYTHON_VENV_DIR}/bin/activate && cd ${E2E_RUNTIME_DIR} && robot ${ROBOT_FLAGS} --pythonpath ${REPO_DIR}/tests ${INCLUDETAG_F} --outputdir ${REPO_DIR}/${E2E_RESULTS_DIR} ${REPO_DIR}/${E2E_TESTS_DIR};deactivate"

clean:
	@rm -fvr ${E2E_RESULTS_DIR} ${E2E_RUNTIME_DIR}
	@rm -fvr ${OUTPUT_DIR}

create-venv:
	sh -c "python3 -m venv ${E2E_PYTHON_VENV_DIR} && . ${E2E_PYTHON_VENV_DIR}/bin/activate && pip install --upgrade pip && pip install -r ${E2E_REQUIREMENTS_FILE} && deactivate"

rm-venv:
	rm -fr ${E2E_PYTHON_VENV_DIR}/

robot-help:
	@sh -c ". ${E2E_PYTHON_VENV_DIR}/bin/activate && robot -h && deactivate"
