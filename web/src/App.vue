<template>
  <div id="app">
    <!-- Bootstrap Navigation Bar -->
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary shadow-sm">
      <div class="container">
        <router-link class="navbar-brand fw-bold" to="/">
          <i class="bi bi-book"></i> Language Learning
        </router-link>
        <button
          class="navbar-toggler"
          type="button"
          data-bs-toggle="collapse"
          data-bs-target="#navbarNav"
          aria-controls="navbarNav"
          aria-expanded="false"
          aria-label="Toggle navigation"
        >
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbarNav">
          <ul class="navbar-nav ms-auto">
            <li class="nav-item">
              <router-link class="nav-link" to="/" active-class="active">
                <i class="bi bi-speedometer2"></i> Dashboard
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/learn" active-class="active">
                <i class="bi bi-book"></i> Learn
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/quiz" active-class="active">
                <i class="bi bi-question-circle"></i> Quiz
              </router-link>
            </li>
          </ul>
        </div>
      </div>
    </nav>

    <!-- Main Content Area -->
    <main class="container my-5">
      <router-view />
    </main>

    <!-- Footer -->
    <footer class="bg-light text-center py-4 mt-5">
      <div class="container">
        <p class="text-muted mb-0">
          &copy; 2024 Language Learning App. Built with Vue.js & Bootstrap
        </p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
</script>

<style>
/* Add Bootstrap icons first */
@import url("https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.1/font/bootstrap-icons.css");

#app {
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

main {
  flex: 1;
}

/* Custom Bootstrap overrides */
.navbar-brand {
  font-size: 1.5rem;
}

.nav-link {
  font-weight: 500;
  transition: all 0.3s ease;
}

.nav-link:hover {
  transform: translateY(-1px);
}

/* Custom animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .navbar-brand {
    font-size: 1.2rem;
  }
}
</style>