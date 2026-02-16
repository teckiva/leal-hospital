# OPD Template - Data Structure & Backend Integration Guide

## Overview
This document describes the exact data structure expected by the OPD template component for rendering patient prescriptions.

## Component Props

The `OPDTemplate` component accepts two main props:

```jsx
<OPDTemplate
  patientData={patientData}
  opdData={opdData}
/>
```

---

## 1. Patient Data Structure

```typescript
interface PatientData {
  name: string;                    // Patient's full name
  age: number | string;            // Patient's age
  sex: string;                     // "Male" | "Female" | "Other"
  address: {
    locality: string;              // Area/locality
    city: string;                  // City name
    state: string;                 // State name
    pincode?: string;              // Postal code (optional)
  };
  date: string;                    // ISO date string or Date object
}
```

### Example:
```json
{
  "name": "Rajesh Kumar",
  "age": 45,
  "sex": "Male",
  "address": {
    "locality": "Swami Vivekananda Nagar",
    "city": "Kushinagar",
    "state": "Uttar Pradesh",
    "pincode": "274401"
  },
  "date": "2026-02-10T10:30:00.000Z"
}
```

---

## 2. OPD Data Structure (The 3 Invisible Parts)

```typescript
interface OPDData {
  // Left Sidebar - Vitals
  vitals: {
    bp?: string;                   // Blood Pressure (e.g., "120/80")
    pulse?: string;                // Pulse rate (e.g., "72 bpm")
    weight?: string;               // Weight (e.g., "65 kg")
    rbs?: string;                  // Random Blood Sugar (e.g., "110 mg/dL")
  };

  // Left Sidebar - Examination
  examination: {
    rightEar?: string;             // Right ear examination notes
    leftEar?: string;              // Left ear examination notes
    ar?: string;                   // Anterior Rhinoscopy findings
    ocOp?: string;                 // Oral Cavity/Oropharynx findings
    vdlFindings?: string;          // Video Digital Laryngoscopy findings
    statusOfNeck?: string;         // Neck examination status
    nodes?: string;                // Lymph nodes examination
  };

  // PART 1: Symptoms (Center content - Invisible section)
  symptoms: string[];              // Array of symptoms

  // PART 2: Prescription/Suggestions (Center content - Invisible section)
  prescription: string[];          // Array of prescription items/suggestions

  // PART 2 (continued): Medicines Table
  medicines: Array<{
    name: string;                  // Medicine name
    volume: string;                // Dosage/volume (e.g., "10ml", "500mg")
    timings: {
      morning: boolean;            // Take in morning
      afternoon: boolean;          // Take in afternoon
      night: boolean;              // Take at night
    };
  }>;

  // PART 3: Future Suggestions (Center content - Invisible section)
  futureSuggestion: string[];      // Array of future suggestions
}
```

### Example:
```json
{
  "vitals": {
    "bp": "120/80",
    "pulse": "72 bpm",
    "weight": "65 kg",
    "rbs": "110 mg/dL"
  },
  "examination": {
    "rightEar": "Normal TM, No discharge",
    "leftEar": "Inflamed TM, Mild discharge",
    "ar": "Deviated septum to right",
    "ocOp": "Mild pharyngitis",
    "vdlFindings": "Normal vocal cords mobility",
    "statusOfNeck": "No lymphadenopathy",
    "nodes": "Not palpable"
  },
  "symptoms": [
    "Ear pain for 3 days",
    "Hearing difficulty in left ear",
    "Occasional headache",
    "Mild fever (99°F)"
  ],
  "prescription": [
    "Clean ear with warm water twice daily",
    "Avoid water entry in ears",
    "Complete the antibiotic course",
    "Use prescribed ear drops as directed"
  ],
  "medicines": [
    {
      "name": "Amoxicillin 500mg",
      "volume": "500mg",
      "timings": {
        "morning": true,
        "afternoon": false,
        "night": true
      }
    },
    {
      "name": "Ciprodex Ear Drops",
      "volume": "3 drops",
      "timings": {
        "morning": true,
        "afternoon": true,
        "night": true
      }
    },
    {
      "name": "Paracetamol",
      "volume": "650mg",
      "timings": {
        "morning": false,
        "afternoon": false,
        "night": true
      }
    }
  ],
  "futureSuggestion": [
    "Follow-up after 5 days",
    "Get audiometry test if hearing issue persists",
    "Avoid cold water and cold foods",
    "Return immediately if pain increases or discharge becomes purulent"
  ]
}
```

---

## 3. The "3 Invisible Parts" Explained

According to the PRD, the center content area contains **3 invisible sections** that only appear when data is present:

### Part 1: SYMPTOMS
- **Location**: Center content, below Rx symbol
- **Purpose**: List all patient symptoms
- **Data Type**: Array of strings
- **Rendering**: Bullet list format
- **Styling**: No visible borders, just content with underlined title

### Part 2: PRESCRIPTION & MEDICINES
- **Location**: Center content, middle section
- **Purpose**:
  - Prescription suggestions/instructions
  - Detailed medicine schedule with timings
- **Data Type**:
  - `prescription`: Array of strings
  - `medicines`: Array of medicine objects
- **Rendering**:
  - Prescription as bullet list
  - Medicines as structured table with morning/afternoon/night columns
- **Styling**: No visible borders for prescription list, bordered table for medicines

### Part 3: FUTURE SUGGESTIONS
- **Location**: Center content, lower section
- **Purpose**: Future medical advice, follow-up instructions, tests required
- **Data Type**: Array of strings
- **Rendering**: Bullet list format
- **Styling**: No visible borders, just content with underlined title

---

## 4. Backend API Response Format

When fetching OPD data from backend, the API should return:

```typescript
interface OPDResponse {
  success: boolean;
  data: {
    patient: PatientData;
    opd: OPDData;
  };
}
```

### Example API Response:
```json
{
  "success": true,
  "data": {
    "patient": {
      "name": "Rajesh Kumar",
      "age": 45,
      "sex": "Male",
      "address": {
        "locality": "Swami Vivekananda Nagar",
        "city": "Kushinagar",
        "state": "Uttar Pradesh",
        "pincode": "274401"
      },
      "date": "2026-02-10T10:30:00.000Z"
    },
    "opd": {
      "vitals": {
        "bp": "120/80",
        "pulse": "72 bpm",
        "weight": "65 kg",
        "rbs": "110 mg/dL"
      },
      "examination": {
        "rightEar": "Normal TM",
        "leftEar": "Inflamed TM",
        "ar": "Deviated septum",
        "ocOp": "Mild pharyngitis",
        "vdlFindings": "Normal",
        "statusOfNeck": "No lymphadenopathy",
        "nodes": "Not palpable"
      },
      "symptoms": [
        "Ear pain for 3 days",
        "Hearing difficulty in left ear"
      ],
      "prescription": [
        "Clean ear with warm water twice daily",
        "Complete the antibiotic course"
      ],
      "medicines": [
        {
          "name": "Amoxicillin 500mg",
          "volume": "500mg",
          "timings": {
            "morning": true,
            "afternoon": false,
            "night": true
          }
        }
      ],
      "futureSuggestion": [
        "Follow-up after 5 days",
        "Get audiometry test if needed"
      ]
    }
  }
}
```

---

## 5. Usage Example in React Component

```jsx
import React, { useState, useEffect } from 'react';
import OPDTemplate from './pages/OPDTemplate/OPDTemplate';

function GenerateOPDPage() {
  const [opdResponse, setOpdResponse] = useState(null);
  const [loading, setLoading] = useState(false);

  const fetchOPD = async (mobileOrOpdId) => {
    setLoading(true);
    try {
      const response = await fetch(`/api/opd/generate?id=${mobileOrOpdId}`);
      const data = await response.json();

      if (data.success) {
        setOpdResponse(data.data);
      }
    } catch (error) {
      console.error('Error fetching OPD:', error);
    } finally {
      setLoading(false);
    }
  };

  const handlePrint = () => {
    window.print();
  };

  return (
    <div>
      {loading && <div>Loading OPD...</div>}

      {opdResponse && (
        <>
          <button onClick={handlePrint}>Print OPD</button>
          <OPDTemplate
            patientData={opdResponse.patient}
            opdData={opdResponse.opd}
          />
        </>
      )}
    </div>
  );
}

export default GenerateOPDPage;
```

---

## 6. Important Notes

### Data Validation:
- All fields are optional (using `?.` optional chaining)
- Empty arrays will render empty space placeholders
- Missing vitals/examination data will show empty fields
- Date formatting is handled automatically (converts to Indian format)

### Print Functionality:
- The template is A4 size (210mm x 297mm)
- Uses `@media print` styles for proper printing
- All colors and layouts are print-optimized
- Page breaks are handled to avoid content splitting

### Responsive Design:
- On screens < 900px, layout switches to mobile-friendly view
- Sidebar appears above content on mobile
- Font sizes adjust for smaller screens

### "Invisible" Sections:
- These sections have no borders or background colors
- They appear as natural content flow
- When printed, they look like handwritten prescription sections
- The watermark and Rx symbol remain in background (low opacity)

---

## 7. Database Schema Alignment (From PRD)

The OPD data structure aligns with the database tables:

### `LaelPatients` table provides:
- name
- age
- sex
- address fields
- mobile
- opdId

### `PatientOpd` table provides:
- symptoms (JSON array)
- prescription (JSON array)
- medicines (JSON array with timing structure)
- futureSuggestion (JSON array)
- vitals and examination data

---

## 8. Testing with Mock Data

For testing purposes, you can use this mock data:

```jsx
const mockPatientData = {
  name: "Test Patient",
  age: 30,
  sex: "Male",
  address: {
    locality: "Test Locality",
    city: "Test City",
    state: "Test State",
    pincode: "123456"
  },
  date: new Date().toISOString()
};

const mockOpdData = {
  vitals: {
    bp: "120/80",
    pulse: "72",
    weight: "70kg",
    rbs: "100"
  },
  examination: {
    rightEar: "Normal",
    leftEar: "Normal",
    ar: "Normal",
    ocOp: "Normal"
  },
  symptoms: ["Symptom 1", "Symptom 2"],
  prescription: ["Instruction 1", "Instruction 2"],
  medicines: [
    {
      name: "Medicine 1",
      volume: "10ml",
      timings: { morning: true, afternoon: false, night: true }
    }
  ],
  futureSuggestion: ["Follow up in 1 week"]
};

<OPDTemplate patientData={mockPatientData} opdData={mockOpdData} />
```

---

## Summary

✅ **Layout matches reference image exactly**
✅ **All fields (Name, Age, Sex, Address, Date) properly bound**
✅ **3 invisible sections implemented (Symptoms, Prescription/Medicines, Future Suggestions)**
✅ **Left sidebar with vitals and examination fields**
✅ **Backend data binding ready**
✅ **Print-ready A4 format**
✅ **Responsive for screen preview**

The template is now ready for backend integration and can display complete OPD prescriptions with all patient data, clinical findings, and treatment plans.
