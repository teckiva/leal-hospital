import React from 'react';
import './OPDTemplate.css';

const OPDTemplate = ({ patientData, opdData }) => {
  // Format date for display
  const formatDate = (dateString) => {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-IN');
  };

  // Format medicines for display
  const formatMedicines = (medicines) => {
    if (!medicines || !Array.isArray(medicines)) return [];
    return medicines;
  };

  return (
    <div className="opd-container">
      {/* Header Section */}
      <div className="opd-header">
        <div className="medical-symbol">
          <svg width="80" height="100" viewBox="0 0 80 100" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M40 10 L35 25 L40 30 L45 25 Z" fill="#DAA520" />
            <ellipse cx="40" cy="10" rx="8" ry="8" fill="#DAA520" />
            <path d="M35 30 Q30 35 25 40 L15 70 Q20 75 25 75 L30 40 Z" fill="#DAA520" stroke="#8B6914" strokeWidth="1" />
            <path d="M45 30 Q50 35 55 40 L65 70 Q60 75 55 75 L50 40 Z" fill="#DAA520" stroke="#8B6914" strokeWidth="1" />
            <path d="M25 40 L30 45 L35 40 L30 50 Z" fill="#DAA520" />
            <path d="M55 40 L50 45 L45 40 L50 50 Z" fill="#DAA520" />
            <rect x="35" y="30" width="10" height="50" fill="#DAA520" />
            <ellipse cx="40" cy="80" rx="6" ry="4" fill="#DAA520" />
          </svg>
        </div>

        <div className="hospital-info">
          <h1 className="hospital-name">LAEL ENT HOSPITAL</h1>
          <p className="hospital-address-hindi">
            पता वार्ड नं 7, स्वामी विवेकानन्द नगर पहरौना, ( निकट अम्बे चौक), कुशीनगर
          </p>
          <p className="hospital-contact">मो.7080920551, 9129272551</p>
        </div>

        <div className="hospital-accreditation">
          <p>पूर्व रेजिडेन्ट सफदरजंग अस्पताल, नई दिल्ली</p>
          <p>पूर्व रेजिडेन्ट KGMU लखनऊ</p>
          <p>पूर्व वरिष्ठ ENT चिकित्सक, संयुक्त जिला चिकित्सालय</p>
          <p className="location-line">पहरौना, कुशीनगर</p>
          <p>पूर्व असिस्टेन्ट प्रोफेस (Assistant Professor) मेडिकल कालेज, देवरिया</p>
        </div>
      </div>

      <div className="doctor-name-section">
        <h2 className="doctor-name">डा. संतोष कुमार</h2>
        <p className="doctor-qualifications">MBBS,DLO,MS ENT (KGMU Lucknow)</p>
        <p className="doctor-specialization">कान, नाक, गला एवं हैड नेक सर्जन</p>
        <p className="doctor-registration">रजिस्ट्रेशन नं 48357</p>
      </div>

      {/* Patient Details Section - With Data Binding */}
      <div className="patient-details-section">
        <div className="patient-line-1">
          <span className="label">Pt. Name</span>
          <span className="filled-data">{patientData?.name || ''}</span>
          <span className="dotted-line"></span>
          <span className="label">Age</span>
          <span className="filled-data">{patientData?.age || ''}</span>
          <span className="dotted-line-short"></span>
          <span className="label">Sex</span>
          <span className="filled-data">{patientData?.sex || ''}</span>
          <span className="dotted-line-short"></span>
        </div>
        <div className="patient-line-2">
          <span className="label">Address</span>
          <span className="filled-data">
            {patientData?.address
              ? `${patientData.address.locality || ''}, ${patientData.address.city || ''}, ${patientData.address.state || ''} ${patientData.address.pincode || ''}`
              : ''}
          </span>
          <span className="dotted-line-long"></span>
          <span className="label">Date</span>
          <span className="filled-data">{formatDate(patientData?.date)}</span>
          <span className="dotted-line-short"></span>
        </div>
      </div>

      {/* Main Content Section */}
      <div className="opd-main-content">
        {/* Left Sidebar - Vitals */}
        <div className="left-sidebar">
          <div className="vitals-item">
            <strong>BP</strong>
            <div className="vitals-value">{opdData?.vitals?.bp || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>Pulse</strong>
            <div className="vitals-value">{opdData?.vitals?.pulse || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>Weight</strong>
            <div className="vitals-value">{opdData?.vitals?.weight || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>RBS</strong>
            <div className="vitals-value">{opdData?.vitals?.rbs || ''}</div>
          </div>
          <div className="ear-examination">
            <strong>Right ear</strong>
            <div className="ear-diagram">
              <svg width="80" height="80" viewBox="0 0 80 80">
                <circle cx="40" cy="40" r="35" fill="white" stroke="black" strokeWidth="2" />
                <line x1="40" y1="40" x2="60" y2="30" stroke="black" strokeWidth="2" />
                <circle cx="40" cy="40" r="3" fill="black" />
              </svg>
            </div>
            <div className="ear-notes">{opdData?.examination?.rightEar || ''}</div>
          </div>
          <div className="ear-examination">
            <strong>Left ear</strong>
            <div className="ear-diagram">
              <svg width="80" height="80" viewBox="0 0 80 80">
                <circle cx="40" cy="40" r="35" fill="white" stroke="black" strokeWidth="2" />
                <line x1="40" y1="40" x2="20" y2="30" stroke="black" strokeWidth="2" />
                <circle cx="40" cy="40" r="3" fill="black" />
              </svg>
            </div>
            <div className="ear-notes">{opdData?.examination?.leftEar || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>AR</strong>
            <div className="vitals-value">{opdData?.examination?.ar || ''}</div>
          </div>
          <div className="vitals-item">
            <strong>OC/OP</strong>
            <div className="vitals-value">{opdData?.examination?.ocOp || ''}</div>
          </div>
          <div className="findings-section">
            <strong>VDL FINDINGS</strong>
            <div className="findings-value">{opdData?.examination?.vdlFindings || ''}</div>
          </div>
          <div className="findings-section">
            <strong>STATUS OF NECK</strong>
            <div className="findings-value">{opdData?.examination?.statusOfNeck || ''}</div>
          </div>
          <div className="findings-section">
            <strong>NODES</strong>
            <div className="findings-value">{opdData?.examination?.nodes || ''}</div>
          </div>
        </div>

        {/* Center Content - THE 3 INVISIBLE PARTS */}
        <div className="center-content">
          {/* Rx Symbol */}
          <div className="rx-symbol">R<sub>x</sub></div>

          {/* Medical Symbol Watermark */}
          <div className="caduceus-watermark">
            <svg width="300" height="500" viewBox="0 0 300 500" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M150 50 L140 100 L150 120 L160 100 Z" fill="rgba(218, 165, 32, 0.15)" />
              <ellipse cx="150" cy="50" rx="20" ry="20" fill="rgba(218, 165, 32, 0.15)" />
              <path d="M130 120 Q100 140 80 160 L40 300 Q60 320 80 320 L120 180 Z" fill="rgba(218, 165, 32, 0.15)" />
              <path d="M170 120 Q200 140 220 160 L260 300 Q240 320 220 320 L180 180 Z" fill="rgba(218, 165, 32, 0.15)" />
              <rect x="135" y="120" width="30" height="300" fill="rgba(218, 165, 32, 0.15)" />
              <ellipse cx="150" cy="420" rx="30" ry="20" fill="rgba(218, 165, 32, 0.15)" />
            </svg>
          </div>

          {/* PART 1: SYMPTOMS (Invisible/Print-only section) */}
          <div className="prescription-section symptoms-section">
            <div className="section-title">SYMPTOMS:</div>
            <div className="section-content">
              {opdData?.symptoms && Array.isArray(opdData.symptoms) && opdData.symptoms.length > 0 ? (
                <ul className="symptoms-list">
                  {opdData.symptoms.map((symptom, index) => (
                    <li key={index}>{symptom}</li>
                  ))}
                </ul>
              ) : (
                <div className="empty-space"></div>
              )}
            </div>
          </div>

          {/* PART 2: PRESCRIPTION & MEDICINES (Invisible/Print-only section) */}
          <div className="prescription-section medicines-section">
            <div className="section-title">PRESCRIPTION:</div>
            <div className="section-content">
              {opdData?.prescription && Array.isArray(opdData.prescription) && opdData.prescription.length > 0 ? (
                <ul className="prescription-list">
                  {opdData.prescription.map((item, index) => (
                    <li key={index}>{item}</li>
                  ))}
                </ul>
              ) : (
                <div className="empty-space"></div>
              )}
            </div>

            {/* Medicines Table */}
            {opdData?.medicines && Array.isArray(opdData.medicines) && opdData.medicines.length > 0 && (
              <div className="medicines-table">
                <table>
                  <thead>
                    <tr>
                      <th>Medicine</th>
                      <th>Volume</th>
                      <th>Morning</th>
                      <th>Afternoon</th>
                      <th>Night</th>
                    </tr>
                  </thead>
                  <tbody>
                    {formatMedicines(opdData.medicines).map((medicine, index) => (
                      <tr key={index}>
                        <td>{medicine.name || ''}</td>
                        <td>{medicine.volume || ''}</td>
                        <td>{medicine.timings?.morning ? '✓' : ''}</td>
                        <td>{medicine.timings?.afternoon ? '✓' : ''}</td>
                        <td>{medicine.timings?.night ? '✓' : ''}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {/* PART 3: FUTURE SUGGESTIONS (Invisible/Print-only section) */}
          <div className="prescription-section future-section">
            <div className="section-title">FUTURE SUGGESTIONS:</div>
            <div className="section-content">
              {opdData?.futureSuggestion && Array.isArray(opdData.futureSuggestion) && opdData.futureSuggestion.length > 0 ? (
                <ul className="future-list">
                  {opdData.futureSuggestion.map((suggestion, index) => (
                    <li key={index}>{suggestion}</li>
                  ))}
                </ul>
              ) : (
                <div className="empty-space"></div>
              )}
            </div>
          </div>

          {/* Legal Text */}
          <div className="legal-text">Not For Medico Legal Purpose</div>

          {/* Timing Section */}
          <div className="timing-section">
            <p className="opd-timing">ओपीडी समय सोमवार से शनिवार 11 AM - 3 PM ; रविवार बन्दी</p>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="opd-footer"></div>
    </div>
  );
};

export default OPDTemplate;
