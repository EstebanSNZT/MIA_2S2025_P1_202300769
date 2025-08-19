<script setup>
// import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="container-fluid">
    <div class="d-flex align-items-center justify-content-center">
      <h1 class="main-title">
        <i class="bi bi-hdd"></i>
        GoDisk
      </h1>
    </div>
    <div class="d-flex align-items-center justify-content-center mb-4">
      <h2 class="subtitle">Sistema de archivos EXT2</h2>
    </div>
    <div class="card border-0 shadow-lg mb-4 bg-white rounded">
      <div class="card-header p-3">
        <div class="d-flex align-items-center">
          <h5 class="mb-0">
            <i class="bi bi-terminal-fill me-2 fs-5"></i>
            Terminal
          </h5>
        </div>
      </div>
      <div class="card-body p-4">
        <div class="form-floating mb-3">
          <textarea v-model="entrada" class="form-control bg-light" id="commandTextarea" style="height: 120px"
            placeholder="Escribe comandos aquí..."></textarea>
          <label for="commandTextarea">Ingresa el comando o script</label>
        </div>

        <div class="row g-3">
          <div class="col-md-6">
            <div class="input-group">
              <label class="input-group-text bg-secondary text-white">
                <i class="bi bi-file-earmark-text"></i>
              </label>
              <input type="file" class="form-control" @change="handleFileUpload" accept=".mias" id="fileInput" />
            </div>
            <div class="form-text text-muted mt-1">
              <i class="bi bi-info-circle-fill me-1"></i> Solo archivos con extensión .mias
            </div>
            <div v-if="fileError" class="alert alert-danger mt-2 py-2 small">
              <i class="bi bi-exclamation-triangle-fill me-1"></i> {{ fileError }}
            </div>
          </div>
          <div class="col-md-3">
            <button class="btn btn-primary w-100 d-flex justify-content-center align-items-center" @click="ejecutar">
              <i class="bi bi-play-fill me-2"></i> Ejecutar
            </button>
          </div>
          <div class="col-md-3">
            <button class="btn btn-danger w-100 d-flex justify-content-center align-items-center" @click="limpiar">
              <i class="bi bi-trash me-2"></i> Limpiar
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container-fluid {
  margin-top: 50px;
  padding: 0px 50px;
}

.main-title {
  font-size: 100px;
  font-weight: bold;
  margin: 0;
  color: white;
  -webkit-text-stroke: 3px #007bff;
  text-shadow: 5px 5px 2px rgba(0, 123, 255, 0.3);
  text-align: center;
  transition: all 0.3s ease;
}

.subtitle {
  font-size: 28px;
  color: #656e75;
  text-align: center;
}

.main-title:hover {
  transform: scale(1.05);
}

.card {
  transition: transform 0.2s;
}

.card:hover {
  transform: translateY(-4px);
}

.card-header {
  background-color: #36a30b;
  color: white;
}

textarea.form-control:focus {
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
  border-color: #86b7fe;
}
</style>
