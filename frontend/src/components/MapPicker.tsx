import { useState } from 'react'
import L from 'leaflet'
import { MapContainer, Marker, TileLayer, useMapEvents } from 'react-leaflet'
import { PinIcon } from './icons'

export interface Point {
  lat: number
  lng: number
}

// Tehran. Only the starting view: the customer clicks the map to drop the pin.
const DEFAULT_CENTER: [number, number] = [35.6892, 51.389]

// A pin drawn with CSS (see .dakeh-pin in index.css), so no image files are needed.
const pinIcon = L.divIcon({ className: 'dakeh-pin', html: '<span></span>', iconSize: [32, 36], iconAnchor: [16, 34] })

const round = (n: number) => Math.round(n * 1e6) / 1e6

function ClickToPick({ onPick }: { onPick: (point: Point) => void }) {
  useMapEvents({
    click(e) {
      onPick({ lat: round(e.latlng.lat), lng: round(e.latlng.lng) })
    },
  })
  return null
}

export function MapPicker({ value, onChange, error }: { value: Point | null; onChange: (point: Point) => void; error?: string }) {
  const [map, setMap] = useState<L.Map | null>(null)
  const [locating, setLocating] = useState(false)
  const [locateError, setLocateError] = useState('')

  function useMyLocation() {
    if (!navigator.geolocation) {
      setLocateError('Your browser cannot share your location. Click the map instead.')
      return
    }
    setLocating(true)
    setLocateError('')
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLocating(false)
        onChange({ lat: round(pos.coords.latitude), lng: round(pos.coords.longitude) })
        map?.setView([pos.coords.latitude, pos.coords.longitude], 16)
      },
      () => {
        setLocating(false)
        setLocateError('We could not get your location. Click the map instead.')
      },
      { enableHighAccuracy: true, timeout: 10000 },
    )
  }

  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between gap-3">
        <span className="text-sm font-semibold">Location on the map</span>
        <button type="button" className="btn btn-quiet btn-sm" onClick={useMyLocation} disabled={locating}>
          <PinIcon className="h-4 w-4" />
          {locating ? 'Finding you…' : 'Use my location'}
        </button>
      </div>

      <div
        className={`h-72 overflow-hidden rounded-lg border-[1.5px] ${error ? 'border-brick' : 'border-mist'}`}
        role="group"
        aria-label="Map. Click to place the pin."
      >
        <MapContainer
          center={value ? [value.lat, value.lng] : DEFAULT_CENTER}
          zoom={value ? 15 : 11}
          scrollWheelZoom
          ref={setMap}
          className="h-full w-full"
        >
          <TileLayer
            attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          />
          <ClickToPick onPick={onChange} />
          {value && (
            <Marker
              position={[value.lat, value.lng]}
              icon={pinIcon}
              draggable
              eventHandlers={{
                dragend: (e) => {
                  const p = (e.target as L.Marker).getLatLng()
                  onChange({ lat: round(p.lat), lng: round(p.lng) })
                },
              }}
            />
          )}
        </MapContainer>
      </div>

      <p className={`mt-1.5 text-sm ${error ? 'text-brick' : 'text-ink/65'}`}>
        {error ??
          (value ? `Pin at ${value.lat.toFixed(6)}, ${value.lng.toFixed(6)}. Drag it or click elsewhere to move it.` : 'Click the map to drop a pin on the address.')}
      </p>
      {locateError && <p className="mt-1 text-sm text-brick">{locateError}</p>}
    </div>
  )
}
