<script>
// import { RouterLink, RouterView } from 'vue-router'

export default {
  data() {
    return {
      input: "",
      output: "",
    }
  },
  methods: {
    handleClean() {
      this.input = "";
      this.output = "";
      const fileInput = document.getElementById('fileInput');
      if (fileInput) fileInput.value = '';
    },
    handleFileUpload(event) {
      const file = event.target.files[0];

      if (!file) return;

      const reader = new FileReader();
      reader.onload = (e) => {
        this.input = e.target.result;
        this.output = "Archivo cargado correctamente. Listo para ejecutar.";
      };
      reader.readAsText(file);
    },
    async handleExecute() {
      if (!this.input.trim()) {
        this.output = "Por favor, ingresa un comando o script.";
        return;
      }

      try {
        const response = await fetch('http://localhost:8000/execute', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ script: this.input })
        });

        const data = await response.json();

        if (data.error) {
          this.output = `Error: ${data.error}`;
        } else if (!response.ok) {
          this.output = `Error HTTP ${response.status}`;
          if (data.output) {
            this.output += ` — ${data.output}`;
          } else if (data.error) {
            this.output += ` — ${data.error}`;
          }
        } else {
          if (data.output && data.output.trim() !== "") {
            this.output = `${data.output}`;
          }
        }
      } catch (error) {
        this.output = `Error de red: ${error.message}`;
      }
    }
  }
}
</script>

<template>
  <div class="container-fluid py-4">
    <div class="text-center mb-4">
      <h1 class="main-title">
        <i class="bi bi-hdd"></i>
        GoDisk
      </h1>
      <p class="subtitle">
        Sistema de archivos EXT2
      </p>
    </div>

    <div class="row justify-content-center">
      <div class="col-lg-12 col-xl-11">
        <div class="card border-0 shadow-lg" style="border-radius: 20px; overflow: hidden;">
          <div class="card-header text-white p-3"
            style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
            <div class="d-flex align-items-center justify-content-between">
              <h4 class="mb-0">
                <i class="bi bi-terminal-fill me-2 fs-4"></i>
                Terminal de comandos
              </h4>
            </div>
          </div>
          <div class="card-body p-4" style="background: linear-gradient(to bottom, #f8f9fa, #ffffff);">
            <div class="form-floating mb-4">
              <textarea v-model="input" class="form-control bg-dark text-light border-2" id="commandTextarea"
                style="height: 140px; font-size: 15px;" placeholder=""></textarea>
              <label class="text-light" for="commandTextarea">
                <i class="bi bi-code me-2"></i>Ingresa el comando o script
              </label>
            </div>

            <div class="row g-3">
              <div class="col-md-5">
                <div class="d-flex flex-column gap-2 h-100">
                  <button class="btn btn-primary flex-fill fw-semibold execute-btn" style="font-size: 1.1rem;"
                    @click="handleExecute">
                    <i class="bi bi-rocket-takeoff me-2"></i>
                    Ejecutar comando
                  </button>
                  <button class="btn btn-outline-secondary flex-fill fw-semibold" style="font-size: 1.1rem;"
                    @click="handleClean">
                    <i class="bi bi-trash3 me-2"></i>
                    Limpiar todo
                  </button>
                </div>
              </div>
              <div class="col-md-7">
                <div class="card border-0 h-100" style="background: #f5f5f5;">
                  <div class="card-body py-2">
                    <div class="d-flex align-items-center h-100">
                      <div class="d-flex flex-column align-items-center text-center" style="min-width: 120px;">
                        <i class="bi bi-file-earmark-plus fs-4 text-primary mb-1"></i>
                        <h6 class="mb-0">Cargar archivo</h6>
                      </div>

                      <div class="vr mx-3" style="height: 60px;"></div>

                      <div class="flex-grow-1">
                        <input type="file" class="form-control mb-2" @change="handleFileUpload" accept=".mias"
                          id="fileInput">
                        <small class="text-muted">
                          <i class="bi bi-info-circle me-1"></i>
                          Solo archivos con extensión .mias
                        </small>
                        <div v-if="fileError" class="alert alert-danger mt-2 py-1 small mb-0">
                          <i class="bi bi-exclamation-triangle-fill me-1"></i>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="row justify-content-center mt-4">
      <div class="col-lg-12 col-xl-11">
        <div class="card border-0 shadow-lg" style="border-radius: 20px; overflow: hidden;">
          <div class="card-header text-white p-3"
            style="background: linear-gradient(135deg, #4a90e2 0%, #1e40af 100%);">
            <div class="d-flex align-items-center">
              <h4 class="mb-0">
                <i class="bi bi-display me-2 fs-4"></i>
                Resultado de la ejecución
              </h4>
            </div>
          </div>
          <div class="card-body p-4" style="background: linear-gradient(to bottom, #f8f9fa, #ffffff);">
            <textarea v-model="output" readonly class="form-control bg-dark text-light" id="resultTextarea"
              style="height: 200px; font-size: 14px; border: 2px solid #333; font-family: 'Courier New', monospace;"></textarea>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container-fluid {
  margin: 20px 0px;
  padding: 0px 30px;
}

.main-title {
  font-size: 95px;
  font-weight: bold;
  margin: -10px 0px 0px 0px;
  padding: 0;
  color: white;
  -webkit-text-stroke: 3px #007bff;
  text-shadow: 5px 5px 2px rgba(0, 123, 255, 0.3);
  text-align: center;
  transition: all 0.3s ease;
}

.subtitle {
  font-size: 25px;
  color: #656e75;
  text-align: center;
  margin: -10px 0 20px 0;
}

.main-title:hover {
  transform: scale(1.03);
}

.card {
  transition: transform 0.2s;
}

.card:hover {
  transform: translateY(-4px);
}

textarea.form-control:focus {
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
  border-color: #86b7fe;
}

.form-floating>.form-control:focus~label,
.form-floating>.form-control:not(:placeholder-shown)~label {
  color: #d4d4d4 !important;
}

.form-floating>.form-control:focus~label::after,
.form-floating>.form-control:not(:placeholder-shown)~label::after {
  background-color: transparent !important;
}

.execute-btn {
  background: linear-gradient(135deg, #a8e470 0%, #4bb85a 100%) !important;
  border: none !important;
  transition: all 0.3s ease;
}

.execute-btn:hover {
  background: linear-gradient(135deg, #96c56b 0%, #328b3e 100%) !important;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}
</style>