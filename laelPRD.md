# laelPRD.md

## 1️⃣ README 1 — Original Project Requirements (Full Detailed Version)

### **Name**
Lael Hospital

### **Purpose**
Hospital internal management system

### **Tech Stack**
- **Backend:** Go  
- **Frontend:** React + Vite  
- **Database:** SQLC  

---

## **Features**

### **Login / Authentication / Authorization**
There will be two user options:
- **Admin**
- **Staffs**

Both have mostly the same features except Admin has additional *superpages*, such as:

- Manage Staff Page  

---

## **Admin Registration**
1. Fields:  
   - Name  
   - Mobile Number  
   - Email (optional)  
   - Designation (Auto‑filled as *Doctor* since admin will be a doctor)  
   - Password (minimum 8 characters, alphanumeric + special character)

2. Registration is only completed after OTP verification.  
   - OTP validity: **1 minute**

---

## **Admin Login**
- Login using **mobile number + OTP**  
- Session expires after **1 hour of inactivity**

---

## **Staff Registration & Login**
- Staff **cannot directly register**.  
  A registration request is sent to the Admin.  
  Only after Admin approval can the staff become successfully registered.

- Same registration fields as Admin except **designation**, which has 3 options:  
  - doctor  
  - nurse  
  - staff  

- Staff Login Options:  
  - Mobile + Password  
  - Mobile + OTP  

---

## **Forgot Password (Staff Only)**
- Reset using mobile OTP

---

# **Main Flow**

## **Home Page Layout**
The home page contains:
- Header  
- Footer  
- Left Sidebar (always open)

### **Left Sidebar Sections**
Two sections:
1. **Profile Details**  
2. **Options**

### **Admin Options**
- **Manage Staff**  
  - Displays list of registered staff  
  - Clicking a staff opens full details  
  - Shows login history  
  - Offers master options:  
    - Change status → Active / Inactive / Temporary Inactive  

- **Patient History**  
  - Displays today’s patients  
  - Search by:  
    - mobile number  
    - name  
    - OPD id  

- **Approve Requests**  
  - Incoming registration requests of staff  

### **Staff Options**
- Patient History  
- View Request  

---

## **Main Landing Page Content**
The main section contains:
- General content about hospital  
- Images  
- Three primary big options:  
  - **Outdoor Patients**  
  - **Indoor Patient**  
  - **Generate OPD**

---

# **Outdoor Patient Module**

There are **two options**:
1. **Basic Details**  
2. **Doctor Check‑In**

---

## **1. Basic Details**
On clicking Basic Details:

A form opens requesting:
- Name  
- Mobile  
- Age  
- Sex  
- Address (locality/area, city, state, pincode optional)  
- Date (calendar input)

On Submit:
- Backend checks DB using mobile number to see if patient exists.

### **New Patient**
- Backend generates a **new OPD ID**
- OPD ID will be displayed as **barcode** on OPD  
- A new entry is created in DB  
- `revisitId = null`

### **Existing Patient**
- Same OPD ID is reused  
- A new row is inserted with a **new revisitId**  

After this:
- Basic details are inserted  
- Next step is Doctor Check‑In  

---

## **2. Doctor Check‑In**
Data is saved to a new table referenced by `patientId`.

Doctor Check‑In contains 4 main sections:

### **1. Symptoms**
- Doctor listens and fills symptoms  
- Should allow adding multiple symptoms (list)

### **2. Prescription / Suggestions**
- Same structure as symptoms  

### **3. Medicines**
A list of medicines with:
- name  
- volume (in ml)  
- timings (checkboxes):  
  - morning  
  - afternoon  
  - night  

### **4. Future Suggestions**
Examples:
- Tests required  
- Next visit date  
- Additional medical advice  

On Submit:
- All data is saved with reference to the *basic detail* record.  
- Prevents duplicated entries.

---

# **Indoor Patient Module**
Currently not implemented.  
Clicking shows: **Feature not available yet**

---

# **OPD Generation**
A fixed OPD template will be used.

Options:
- Generate Latest OPD  
- Generate Older OPD  

Input:
- Mobile number OR OPD ID

OPD is generated in real‑time from stored DB data.

---

## **Template Design**
A mid black section containing 4 parts:
- 2 columns:  
  1. **Details**  
     - Symptoms  
     - Prescription  
     - Future Suggestion  
  2. **Medicines**

---

# **Data Modeling**

### **1. LaelUsers**
Fields:
- id (auto-increment)  
- name  
- mobile  
- email  
- designation  
- status  
- isAdmin  
- isApproved (always 1 for admin)  
- password  
- createdOn  
- updatedOn  

---

### **2. LaelOtp**
Fields:
- id  
- mobile  
- otp  
- expiry  
- isValidated  
- createdOn  
- updatedOn  

---

### **3. LaelPatients**
Fields:
- id  
- name  
- mobile  
- opdId  
- age  
- sex  
- address  
- revisitId  
- createdOn  
- updatedOn  

---

### **4. PatientOpd**
Fields:
- id  
- patientId  
- symptoms  
- prescription  
- medicines  
- futureSuggestion  
- createdOn  
- updatedOn  

---

---

# **2️⃣ Improvements to Existing Features + Improved Data Modelling (No New Features Added)**

## **Authentication Improvements**
- Add `retryCount` to OTP attempts  
- Add `otpType` (registration/login/forgot-password)  
- Add `sessionToken`, `sessionExpiry`, `lastActiveAt` for better session tracking  

---

## **Staff Approval Flow**
- Add `requestedBy` and `approvedBy` fields for traceability  

---

## **Patient Registration Flow Improvements**
- Replace `revisitId` with a clearer `visitNumber`  
- Use structured address fields: locality, city, state, pincode  
- Use standardized OPD ID format  

---

## **Doctor Check‑In Improvements**
- Store symptoms, prescriptions, medicines, future suggestions as JSON arrays  
- Store medicines as structured JSON with timing flags  
- Add `doctorId` (FK)  
- Add `templateVersion` to ensure consistency if template changes later  

---

# **Improved Data Models**

## **LaelUsers (Improved)**
```
id
name
mobile
email
designation
status
isAdmin
isApproved
approvedBy
passwordHash
createdOn
updatedOn
lastLoginAt
```

---

## **LaelOtp (Improved)**
```
id
mobile
otp
expiry
isValidated
otpType
retryCount
createdOn
updatedOn
```

---

## **LaelPatients (Improved)**
```
id
name
mobile
opdId
age
sex
addressLocality
addressCity
addressState
addressPincode
visitNumber
createdOn
updatedOn
```

---

## **PatientOpd (Improved)**
```
id
patientId
doctorId
symptoms (JSON)
prescription (JSON)
medicines (JSON)
futureSuggestion (JSON)
templateVersion
createdOn
updatedOn
```

---

---

# **3️⃣ README with Extra Basic Features (Optional Small Enhancements)**

### **Basic Extra Features (Not Advanced HMS Features)**

#### **1. Basic Activity Logs**
Track:
- Login  
- Logout  
- Patient creation  
- Staff approval events  

#### **2. Basic UI Notifications**
Show toast messages:
- Success  
- Error  
- OTP expired  
- Invalid credentials  

#### **3. Ability to Export OPD as PDF**
- Download button  
- Uses same fixed template  

#### **4. Basic Search Enhancements**
- Add filters on patient history search  
- Search by date range  

#### **5. Staff Profile Page**
- View personal details  
- Change password  
- View login history  

#### **6. Simple Dark/Light Mode Toggle**
User preference stored in DB or LocalStorage.

---

# End of Document
