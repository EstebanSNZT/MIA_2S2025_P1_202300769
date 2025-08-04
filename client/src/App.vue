<script setup>
// import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="container-fluid">
    <div class="title-container">
      <h1 class="main-title">
        <i class="bi bi-hdd"></i>
        GoDisk
      </h1>
    </div>
    <div class="subtitle-container">
      <h2 class="subtitle">Sistema de archivos EXT2</h2>
    </div>
  </div>
</template>

<style scoped>
.container-fluid {
  margin-top: 50px;
  padding: 0px 50px;
}

.title-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
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

.subtitle-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  margin-bottom: 30px;
}

.subtitle {
  font-size: 28px;
  color: #656e75;
  text-align: center;
}

.main-title:hover {
  transform: scale(1.05);
}
</style>
