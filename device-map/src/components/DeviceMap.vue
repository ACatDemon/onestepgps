<template>
  <div class="map-container">
    <div class="sidebar">
      <div class="top">
        <h2 class="title">Devices</h2>
        <select v-model="sortOption" class="dropdown" @change="updatePreferences">
          <option value="a-z">A to Z</option>
          <option value="z-a">Z to A</option>
          <option value="active">By Status</option>
        </select>
      </div>
      
      <div class="devices">
        <div v-if="loading">
          Loading...
        </div>

        <template v-else>
          <div v-for="device in sortedDevices" 
               :key="device.device_id" 
               class="device">
            <div class="main-content">
              <div class="device-top">
                <div class="dot" :class="{ active: device.active_state === 'active' }" />
                <h3>{{ device.display_name }}</h3>
              </div>
              <div class="state">{{ device.active_state }}</div>
            </div>
            <button @click="toggleHighlight(device.device_id)" 
                    :class="{ blue: highlightedDevices.includes(device.device_id) }">
              Select
            </button>
          </div>
        </template>
      </div>
    </div>

    <div class="map-wrapper">
      <GMapMap v-if="initialCenter"
               :center="initialCenter"
               :zoom="8"
               class="map"
               ref="mapRef">

        <template v-if="!loading">
          <GMapMarker v-for="device in sortedDevices"
                     :key="device.device_id"
                     :position="getDevicePosition(device)"
                     :title="device.display_name"
                     :icon="getMarkerIcon(device)"
                     @click="openInfoWindow(device)">

            <GMapInfoWindow v-if="selectedDevice === device.device_id"
                          :opened="true"
                          @closeclick="selectedDevice = null">
              <div class="info">
                <div class="info-header">
                  <span class="name">{{ device.display_name }}</span>
                  <button @click="toggleHighlight(device.device_id)"
                          :class="{ blue: highlightedDevices.includes(device.device_id) }">
                    Select
                  </button>
                </div>

                <div class="details">
                  <p>Status: {{ device.active_state }}</p>
                  <p class="coords">
                    {{ device.latest_device_point?.lat.toFixed(6) }}, 
                    {{ device.latest_device_point?.lng.toFixed(6) }}
                  </p>
                </div>
              </div>
            </GMapInfoWindow>
          </GMapMarker>
        </template>
      </GMapMap>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue'

export default {
  name: 'DeviceMap',
  setup() {
    const loading = ref(true)
    const devices = ref([])

    async function loadDevices() {
      loading.value = true

      try {
        const response = await fetch('http://localhost:3000/api/devices')
        const data = await response.json()
        devices.value = data.result_list || []
      } catch (err) {
        console.error('Failed to load devices:', err)
      }

      loading.value = false
    }

    const mapRef = ref(null)
    const initialCenter = ref({ lat: 34, lng: -118 })

    function getMarkerIcon(device) {
      if (!window?.google?.maps) {
        return null 
      }
      
      const highlighted = highlightedDevices.value.includes(device.device_id)
      
      return {
        path: window.google.maps.SymbolPath.CIRCLE,
        scale: highlighted ? 10 : 8,
        fillColor: 'red',
        fillOpacity: highlighted ? 1 : 0.7,
        strokeWeight: highlighted ? 3 : 1,
        strokeColor: highlighted ? 'DodgerBlue' : '#000000'
      }
    }

    function getDevicePosition(device) {
      return {
        lat: device.latest_device_point?.lat || initialCenter.value.lat,
        lng: device.latest_device_point?.lng || initialCenter.value.lng
      }
    }

    const selectedDevice = ref(null)
    const highlightedDevices = ref([])

    async function toggleHighlight(id) {
      if (highlightedDevices.value.includes(id)) {
        highlightedDevices.value = highlightedDevices.value.filter(x => x !== id)
      } else {
        highlightedDevices.value.push(id)
      }
      await updatePreferences()
    }

    const sortOption = ref('a-z')
    
    const sortedDevices = computed(() => {
      if (!devices.value?.length) return []
      
      let sorted = [...devices.value]

      if (sortOption.value === 'a-z') {
        return sorted.sort((a, b) => (a.display_name > b.display_name ? 1 : -1))
      }

      if (sortOption.value === 'z-a') {
        return sorted.sort((a, b) => (a.display_name > b.display_name ? -1 : 1))
      }

      return sorted.sort((a, b) => {
        if (a.active_state === b.active_state) {
          return (a.display_name > b.display_name ? 1 : -1)
        }
        return a.active_state === 'active' ? -1 : 1
      })
    })

    async function loadPreferences() {
      try {
        const response = await fetch('http://localhost:3000/api/preferences')
        const data = await response.json()
        sortOption.value = data.sort || 'a-z'
        highlightedDevices.value = data.highlight || []
      } catch (err) {
        console.error('Failed to load preferences:', err)
      }
    }

    async function updatePreferences() {
      try {
        await fetch('http://localhost:3000/api/preferences', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            sort: sortOption.value,
            highlight: highlightedDevices.value
          })
        })
      } catch (err) {
        console.error('Failed to save preferences:', err)
      }
    }

    let refreshInterval

    onMounted(async () => {
      await Promise.all([
        loadPreferences(),
        loadDevices()
      ])

      refreshInterval = setInterval(loadDevices, 30000)
    })

    onUnmounted(() => {
      if (refreshInterval) {
        clearInterval(refreshInterval)
      }
    })

    return {
      loading,
      devices,
      sortedDevices,
      mapRef,
      initialCenter,
      sortOption,
      selectedDevice,
      highlightedDevices,
      getMarkerIcon,
      getDevicePosition,
      toggleHighlight,
      openInfoWindow: (device) => selectedDevice.value = device.device_id,
      updatePreferences
    }
  }
}
</script>

<style scoped>
.map-container {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 320px;
  border-right: 1px solid #ddd;
  background: #eee;
  display: flex;
  flex-direction: column;
}

.top {
  padding: 15px;
  background: white;
  border-bottom: 1px solid #ddd;
}

.title {
  font-size: 20px;
  font-weight: bold;
}

.dropdown {
  width: 100%;
  margin-top: 8px;
  padding: 6px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.devices {
  flex-grow: 1;
  overflow-y: auto;
  padding: 10px;
}

.device {
  padding: 12px;
  margin-bottom: 8px;
  background: white;
  border: 1px solid #ddd;
  display: flex;
  align-items: center;
  gap: 8px;
}

.main-content {
  flex-grow: 1;
}

.device-top {
  display: flex;
  align-items: center;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ef4444;
  margin-right: 8px;
}

.dot.active {
  background: #22c55e;
}

.state {
  color: #666;
  font-size: 14px;
  margin-left: 16px;
}

button {
  padding: 6px 12px;
  border-radius: 4px;
  background: #efefef;
}

button.blue {
  background: DodgerBlue;
  color: white;
}

.map-wrapper {
  flex-grow: 1;
}

.map {
  height: 100%;
}

.info {
  min-width: 200px;
  padding: 8px;
}

.info-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.name {
  font-weight: bold;
}

.coords {
  font-family: monospace;
  color: #666;
}
</style>