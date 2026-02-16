import { useState } from 'react'
import './App.css'
import Header from './components/Header/Header'
import OPDTemplate from './pages/OPDTemplate/OPDTemplate'
import { AuthProvider } from './authentication/AuthContext'

function App() {
  const [showOPD, setShowOPD] = useState(false);

  // Example patient and OPD data
  const samplePatientData = {
    name: "Sample Patient",
    age: 35,
    sex: "Male",
    address: "Sample Address, City, State",
    date: new Date().toLocaleDateString()
  };

  const sampleOPDData = {
    bp: "120/80",
    pulse: "72",
    weight: "70kg",
    symptoms: ["Headache", "Fever"],
    prescription: ["Rest", "Drink plenty of water"],
    medicines: [
      { name: "Paracetamol", volume: "500mg", timings: { morning: true, afternoon: false, night: true } }
    ]
  };

  return (
    <AuthProvider>
      <div className="app">
        <Header userName="Dr. Santosh Kumar" userRole="Admin" />

        <div className="app-content">
          <div className="welcome-section">
            <h1>Welcome to Lael Hospital Management System</h1>
            <p>Frontend folder structure has been set up successfully!</p>

            <div className="structure-info">
              <h2>Folder Structure:</h2>
              <ul>
                <li><strong>pages/</strong> - Page components (e.g., OPDTemplate, Login, Dashboard)</li>
                <li><strong>components/</strong> - Reusable components (e.g., Header, Sidebar, Footer)</li>
                <li><strong>authentication/</strong> - Auth context and related files</li>
              </ul>

              <p>Each page/component has its own folder with JSX and CSS files using the same name.</p>
            </div>

            <button
              className="demo-btn"
              onClick={() => setShowOPD(!showOPD)}
            >
              {showOPD ? 'Hide OPD Template' : 'Show OPD Template Demo'}
            </button>
          </div>

          {showOPD && (
            <div className="opd-demo-section">
              <h2>OPD Template Preview</h2>
              <OPDTemplate
                patientData={samplePatientData}
                opdData={sampleOPDData}
              />
            </div>
          )}
        </div>
      </div>
    </AuthProvider>
  )
}

export default App
